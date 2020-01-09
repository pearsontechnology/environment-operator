package translator

// translator package converts objects between Kubernetes and Bitesize

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
	autoscale_v2beta1 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	v1beta1_ext "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KubeMapper maps BitesizeService object to Kubernetes objects
type KubeMapper struct {
	BiteService *bitesize.Service
	Gists       *bitesize.Gists
	Namespace   string
	Config      struct {
		Project        string
		DockerRegistry string
	}
}

// Service extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) Service() (*v1.Service, error) {
	targetServiceName := w.BiteService.Name
	if w.BiteService.IsBlueGreenParentDeployment() {
		targetServiceName = w.BiteService.ActiveDeploymentName()
	}
	var ports []v1.ServicePort
	for _, p := range w.BiteService.Ports {
		servicePort := v1.ServicePort{
			Port:       int32(p),
			TargetPort: intstr.FromInt(p),
			Name:       fmt.Sprintf("tcp-port-%d", p),
		}
		ports = append(ports, servicePort)
	}
	retval := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        w.BiteService.Name,
			Namespace:   w.Namespace,
			Labels:      w.labels(),
			Annotations: w.annotations(),
		},
		Spec: v1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"creator": "pipeline",
				"name":    targetServiceName,
			},
		},
	}
	return retval, nil
}

// HeadlessService extracts Kubernetes Headless Service object (No ClusterIP) from Bitesize definition
func (w *KubeMapper) HeadlessService() (*v1.Service, error) {
	targetServiceName := w.BiteService.Name
	if w.BiteService.IsBlueGreenParentDeployment() {
		targetServiceName = w.BiteService.ActiveDeploymentName()
	}

	var ports []v1.ServicePort
	//Need to update this to have an option to create the headless service (no loadbalancing with Cluster IP not getting set)
	for _, p := range w.BiteService.Ports {
		servicePort := v1.ServicePort{
			Port:       int32(p),
			TargetPort: intstr.FromInt(p),
			Name:       fmt.Sprintf("tcp-port-%d", p),
		}
		ports = append(ports, servicePort)
	}
	retval := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
			},
			Annotations: w.annotations(),
		},
		Spec: v1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"creator": "pipeline",
				"name":    targetServiceName,
			},
			ClusterIP: v1.ClusterIPNone,
		},
	}
	return retval, nil
}

// ConfigMaps returns a list of ConfigMaps defined in the service
// definition
func (w *KubeMapper) ConfigMaps() ([]v1.ConfigMap, error) {
	var retval []v1.ConfigMap

	for _, vol := range w.BiteService.Volumes {
		if vol.IsConfigMapVolume() {
			c := w.Gists.FindByName(vol.Name, bitesize.TypeConfigMap)
			if c != nil {
				retval = append(retval, c.ConfigMap)
			}
		}
	}

	if w.BiteService.InitContainers != nil {
		for _, container := range *w.BiteService.InitContainers {
			for _, vol := range container.Volumes {
				if vol.IsConfigMapVolume() {
					c := w.Gists.FindByName(vol.Name, bitesize.TypeConfigMap)
					if c != nil {
						retval = append(retval, c.ConfigMap)
					}
				}
			}
		}
	}

	return retval, nil
}

// PersistentVolumeClaims returns a list of claims for a BiteService
func (w *KubeMapper) PersistentVolumeClaims() ([]v1.PersistentVolumeClaim, error) {
	var retval []v1.PersistentVolumeClaim

	for _, vol := range w.BiteService.Volumes {
		//Create a PVC only if the volume is not coming from a secret or ConfigMap
		if vol.IsSecretVolume() || vol.IsConfigMapVolume() {
			continue
		}

		ret := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      vol.Name,
				Namespace: w.Namespace,
				Labels: map[string]string{
					"creator":    "pipeline",
					"deployment": w.BiteService.Name,
					"mount_path": strings.Replace(vol.Path, "/", "2F", -1),
					"size":       vol.Size,
					"type":       strings.ToLower(vol.Type),
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: getAccessModesFromString(vol.Modes),
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse(vol.Size),
					},
				},
			},
		}
		if vol.HasManualProvisioning() {
			ret.Spec.VolumeName = vol.Name
			ret.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": vol.Name,
				},
			}
		} else {
			ret.ObjectMeta.Annotations = map[string]string{
				"volume.beta.kubernetes.io/storage-class": "aws-" + strings.ToLower(vol.Type),
			}
		}

		retval = append(retval, ret)
	}
	return retval, nil
}

// Deployment extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) Deployment() (*v1beta1_ext.Deployment, error) {
	if w.BiteService.IsBlueGreenParentDeployment() {
		return nil, nil
	}
	replicas := int32(w.BiteService.Replicas)
	container, err := w.container()
	initContainers, _ := w.initContainers()

	if err != nil {
		return nil, err
	}
	if w.BiteService.Version != "" {
		container.Image = util.Image(w.BiteService.Application, w.BiteService.Version)
	}

	imagePullSecrets, err := w.imagePullSecrets()
	volumes, err := w.volumes()
	if err != nil {
		return nil, err
	}

	retval := &v1beta1_ext.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
				"version":     w.BiteService.Version,
			},
		},
		Spec: v1beta1_ext.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"creator": "pipeline",
					"name":    w.BiteService.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      w.BiteService.Name,
					Namespace: w.Namespace,
					Labels: map[string]string{
						"creator":     "pipeline",
						"application": w.BiteService.Application,
						"name":        w.BiteService.Name,
						"version":     w.BiteService.Version,
					},
					Annotations: w.BiteService.Annotations,
				},
				Spec: v1.PodSpec{
					NodeSelector:     map[string]string{"role": "minion"},
					Containers:       []v1.Container{*container},
					ImagePullSecrets: imagePullSecrets,
					Volumes:          volumes,
					InitContainers:   initContainers,
				},
			},
		},
	}

	return retval, nil
}
func (w *KubeMapper) imagePullSecrets() ([]v1.LocalObjectReference, error) {
	var retval []v1.LocalObjectReference

	pullSecrets := util.RegistrySecrets()

	if pullSecrets != "" {
		result := strings.Split(util.RegistrySecrets(), ",")
		for i := range result {
			var namevalue v1.LocalObjectReference
			namevalue = v1.LocalObjectReference{
				Name: result[i],
			}
			retval = append(retval, namevalue)
		}
	}

	return retval, nil
}

// HPA extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) HPA() (*autoscale_v2beta1.HorizontalPodAutoscaler, error) {
	if w.BiteService.IsBlueGreenParentDeployment() {
		return nil, nil
	}
	retval := &autoscale_v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels:    w.labels(),
		},
		Spec: autoscale_v2beta1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v2beta1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       w.BiteService.Name,
				APIVersion: "extensions/v1beta1",
			},
			MinReplicas: &w.BiteService.HPA.MinReplicas,
			MaxReplicas: w.BiteService.HPA.MaxReplicas,
			Metrics:     w.getMetricSpec(),
		},
	}

	return retval, nil
}

func (w *KubeMapper) getMetricSpec() (m []autoscale_v2beta1.MetricSpec) {
	if w.BiteService.HPA.Metric.Name == "cpu" || w.BiteService.HPA.Metric.Name == "memory" {
		if w.BiteService.HPA.Metric.Name == "cpu" && w.BiteService.HPA.Metric.TargetAverageUtilization != 0 {
			m = append(m, autoscale_v2beta1.MetricSpec{
				Type: autoscale_v2beta1.ResourceMetricSourceType,
				Resource: &autoscale_v2beta1.ResourceMetricSource{
					TargetAverageUtilization: &w.BiteService.HPA.Metric.TargetAverageUtilization,
					Name:                     "cpu",
				},
			},
			)
		}
		if w.BiteService.HPA.Metric.Name == "memory" && w.BiteService.HPA.Metric.TargetAverageUtilization != 0 {
			m = append(m, autoscale_v2beta1.MetricSpec{
				Type: autoscale_v2beta1.ResourceMetricSourceType,
				Resource: &autoscale_v2beta1.ResourceMetricSource{

					TargetAverageUtilization: &w.BiteService.HPA.Metric.TargetAverageUtilization,
					Name:                     "memory",
				},
			},
			)
		}
	} else {
		targetValue, _ := resource.ParseQuantity(w.BiteService.HPA.Metric.TargetAverageValue)
		m = append(m, autoscale_v2beta1.MetricSpec{
			Type: autoscale_v2beta1.PodsMetricSourceType,
			Pods: &autoscale_v2beta1.PodsMetricSource{
				TargetAverageValue: targetValue,
				MetricName:         w.BiteService.HPA.Metric.Name,
			},
		},
		)
	}

	return
}

func (w *KubeMapper) initContainers() ([]v1.Container, error) {
	var retval []v1.Container
	// TODO: Need to add volume, env and other configs support here

	if w.BiteService.InitContainers == nil {
		return nil, nil
	}

	for _, container := range *w.BiteService.InitContainers {
		evars, err := w.initEnvVars(container)
		if err != nil {
			return nil, err
		}

		mounts, err := w.initVolumeMounts(container)
		if err != nil {
			return nil, err
		}

		con := v1.Container{
			Name:         container.Name,
			Image:        "",
			Env:          evars,
			Command:      container.Command,
			VolumeMounts: mounts,
		}

		if container.Version != "" {
			con.Image = util.Image(container.Application, container.Version)
		}

		retval = append(retval, con)
	}

	return retval, nil
}

func (w *KubeMapper) container() (*v1.Container, error) {

	var retval *v1.Container

	mounts, err := w.volumeMounts()
	if err != nil {
		return nil, err
	}

	evars, err := w.envVars()
	if err != nil {
		return nil, err
	}

	resources, err := w.resources()
	if err != nil {
		return nil, err
	}

	liveness, err := w.livenessProbe()
	if err != nil {
		return nil, err
	}

	readiness, err := w.readinessProbe()
	if err != nil {
		return nil, err
	}

	retval = &v1.Container{
		Name:           w.BiteService.Name,
		Image:          "",
		Env:            evars,
		VolumeMounts:   mounts,
		Resources:      resources,
		Command:        w.BiteService.Commands,
		LivenessProbe:  liveness,
		ReadinessProbe: readiness,
	}

	return retval, nil
}

func convertProbeType(probe *bitesize.Probe) *v1.Probe {
	var retval *v1.Probe

	if probe != nil {

		retval = &v1.Probe{
			FailureThreshold:    probe.FailureThreshold,
			InitialDelaySeconds: probe.InitialDelaySeconds,
			SuccessThreshold:    probe.SuccessThreshold,
			TimeoutSeconds:      probe.TimeoutSeconds,
			PeriodSeconds:       probe.PeriodSeconds,
		}

		if probe.HTTPGet != nil {
			httpGet := &v1.HTTPGetAction{}

			for _, v := range probe.HTTPGet.HTTPHeaders {
				httpGet.HTTPHeaders = append(httpGet.HTTPHeaders, v1.HTTPHeader{
					Name:  v.Name,
					Value: v.Value,
				})
			}

			httpGet.Host = probe.HTTPGet.Host
			httpGet.Path = probe.HTTPGet.Path
			httpGet.Port.IntVal = probe.HTTPGet.Port
			httpGet.Scheme = probe.HTTPGet.Scheme

			retval.Handler = v1.Handler{
				HTTPGet: httpGet,
			}
		}

		if probe.Exec != nil {
			exec := &v1.ExecAction{}

			exec.Command = probe.Exec.Command

			retval.Handler = v1.Handler{
				Exec: exec,
			}
		}

		if probe.TCPSocket != nil {
			socket := &v1.TCPSocketAction{}

			socket.Host = probe.TCPSocket.Host
			socket.Port.IntVal = probe.TCPSocket.Port

			retval.Handler = v1.Handler{
				TCPSocket: socket,
			}
		}
	}

	return retval
}

func (w *KubeMapper) livenessProbe() (*v1.Probe, error) {

	probe := w.BiteService.LivenessProbe

	return convertProbeType(probe), nil
}

func (w *KubeMapper) readinessProbe() (*v1.Probe, error) {

	probe := w.BiteService.ReadinessProbe

	return convertProbeType(probe), nil
}

func (w *KubeMapper) initEnvVars(container bitesize.Container) ([]v1.EnvVar, error) {
	var retval []v1.EnvVar
	var err error
	//Create in cluster rest client to be utilized for secrets processing
	client, _ := k8s.ClientForNamespace(config.Env.Namespace)

	for _, e := range container.EnvVars {
		var evar v1.EnvVar
		switch {
		case e.Secret != "":
			kv := strings.Split(e.Value, "/")
			secretName := ""
			secretDataKey := ""

			if len(kv) == 2 {
				secretName = kv[0]
				secretDataKey = kv[1]
			} else {
				secretName = kv[0]
				secretDataKey = secretName
			}

			if !client.Secret().Exists(secretName) {
				log.Debugf("Unable to find Secret %s", secretName)
				err = fmt.Errorf("Unable to find secret [%s] in namespace [%s] when processing envvars for init containers [%s]", secretName, config.Env.Namespace, w.BiteService.Name)
			}

			evar = v1.EnvVar{
				Name: e.Secret,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretDataKey,
					},
				},
			}
		case e.Value != "":
			evar = v1.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			}
		case e.PodField != "":
			evar = v1.EnvVar{
				Name: e.Name,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: e.PodField,
					},
				},
			}
		}
		retval = append(retval, evar)
	}

	if w.BiteService.IsBlueGreenChildDeployment() {
		evar := v1.EnvVar{
			Name:  "POD_DEPLOYMENT_COLOUR",
			Value: w.BiteService.Deployment.BlueGreen.DeploymentColour.String(),
		}
		retval = append(retval, evar)
	}
	return retval, err
}

func (w *KubeMapper) envVars() ([]v1.EnvVar, error) {
	var retval []v1.EnvVar
	var err error
	//Create in cluster rest client to be utilized for secrets processing
	client, _ := k8s.ClientForNamespace(config.Env.Namespace)

	for _, e := range w.BiteService.EnvVars {
		var evar v1.EnvVar
		switch {
		case e.Secret != "":
			kv := strings.Split(e.Value, "/")
			secretName := ""
			secretDataKey := ""

			if len(kv) == 2 {
				secretName = kv[0]
				secretDataKey = kv[1]
			} else {
				secretName = kv[0]
				secretDataKey = secretName
			}

			if !client.Secret().Exists(secretName) {
				log.Debugf("Unable to find Secret %s", secretName)
				err = fmt.Errorf("unable to find secret [%s] in namespace [%s] when processing envvars for deployment [%s]", secretName, config.Env.Namespace, w.BiteService.Name)
			}

			evar = v1.EnvVar{
				Name: e.Secret,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretDataKey,
					},
				},
			}
		case e.Value != "":
			evar = v1.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			}
		case e.PodField != "":
			evar = v1.EnvVar{
				Name: e.Name,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: e.PodField,
					},
				},
			}
		}
		retval = append(retval, evar)
	}

	if w.BiteService.IsBlueGreenChildDeployment() {
		evar := v1.EnvVar{
			Name:  "POD_DEPLOYMENT_COLOUR",
			Value: w.BiteService.Deployment.BlueGreen.DeploymentColour.String(),
		}
		retval = append(retval, evar)
	}
	return retval, err
}

func (w *KubeMapper) initVolumeMounts(container bitesize.Container) ([]v1.VolumeMount, error) {
	var retval []v1.VolumeMount

	if w.BiteService.IsBlueGreenParentDeployment() {
		return retval, nil
	}

	for _, v := range container.Volumes {
		if v.Name == "" || v.Path == "" {
			return nil, fmt.Errorf("volume must have both name and path set")
		}
		vol := v1.VolumeMount{
			Name:      v.Name,
			MountPath: v.Path,
		}
		retval = append(retval, vol)
	}
	return retval, nil
}

func (w *KubeMapper) volumeMounts() ([]v1.VolumeMount, error) {
	var retval []v1.VolumeMount

	if w.BiteService.IsBlueGreenParentDeployment() {
		return retval, nil
	}

	for _, v := range w.BiteService.Volumes {
		if v.Name == "" || v.Path == "" {
			return nil, fmt.Errorf("volume must have both name and path set")
		}
		vol := v1.VolumeMount{
			Name:      v.Name,
			MountPath: v.Path,
		}
		retval = append(retval, vol)
	}
	return retval, nil
}

func (w *KubeMapper) volumes() ([]v1.Volume, error) {
	var retval []v1.Volume
	for _, v := range w.BiteService.Volumes {
		vol := v1.Volume{
			Name:         v.Name,
			VolumeSource: w.volumeSource(v),
		}
		retval = append(retval, vol)
	}

	return retval, nil
}

func (w *KubeMapper) volumeSource(vol bitesize.Volume) v1.VolumeSource {
	if vol.IsSecretVolume() {
		return v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: vol.Name},
		}
	}

	if vol.IsConfigMapVolume() {

		var items []v1.KeyToPath

		for _, v := range vol.Items {
			items = append(items, v1.KeyToPath{
				Key:  v.Key,
				Path: v.Path,
				Mode: v.Mode,
			})
		}

		return v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: vol.Name,
				},
				Items: items,
			},
		}
	}

	return v1.VolumeSource{
		PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: vol.Name},
	}

}

// Ingress extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) Ingress() (*v1beta1_ext.Ingress, error) {
	labels := map[string]string{
		"creator":     "pipeline",
		"application": w.BiteService.Application,
		"name":        w.BiteService.Name,
	}

	if w.BiteService.Ssl != "" {
		labels["ssl"] = w.BiteService.Ssl
	}

	if w.BiteService.HTTPSBackend != "" {
		labels["httpsBackend"] = w.BiteService.HTTPSBackend
	}

	if w.BiteService.HTTPSOnly != "" {
		labels["httpsOnly"] = w.BiteService.HTTPSOnly
	}

	if w.BiteService.HTTP2 != "" {
		labels["http2"] = w.BiteService.HTTP2
	}

	port := intstr.FromInt(w.BiteService.Ports[0])
	retval := &v1beta1_ext.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1_ext.IngressSpec{
			Rules: []v1beta1_ext.IngressRule{},
		},
	}

	for _, url := range w.BiteService.ExternalURL {
		rule := v1beta1_ext.IngressRule{
			Host: url,
			IngressRuleValue: v1beta1_ext.IngressRuleValue{
				HTTP: &v1beta1_ext.HTTPIngressRuleValue{
					Paths: []v1beta1_ext.HTTPIngressPath{
						{
							Path: "/",
							Backend: v1beta1_ext.IngressBackend{
								ServiceName: w.BiteService.Name,
								ServicePort: port,
							},
						},
					},
				},
			},
		}

		if w.BiteService.Ssl == "true" {
			tls := v1beta1_ext.IngressTLS{
				Hosts:      []string{url},
				SecretName: w.BiteService.Name,
			}
			retval.Spec.TLS = []v1beta1_ext.IngressTLS{tls}
		} else {
			retval.Spec.TLS = nil
		}

		// Override backend
		if w.BiteService.Backend != "" {
			rule.IngressRuleValue.HTTP.Paths[0].Backend.ServiceName = w.BiteService.Backend
		}
		if w.BiteService.BackendPort != 0 {
			rule.IngressRuleValue.HTTP.Paths[0].Backend.ServicePort = intstr.FromInt(w.BiteService.BackendPort)
		}
		retval.Spec.Rules = append(retval.Spec.Rules, rule)

	}

	return retval, nil
}

// CustomResourceDefinition extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) CustomResourceDefinition() (*ext.PrsnExternalResource, error) {
	retval := &ext.PrsnExternalResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       strings.Title(w.BiteService.Type),
			APIVersion: getAPIVersion(w.BiteService.Options),
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
			Namespace:       w.Namespace,
			Name:            w.BiteService.Name,
			ResourceVersion: w.BiteService.ResourceVersion,
		},
		Spec: ext.PrsnExternalResourceSpec{
			Version:         w.BiteService.Version,
			Options:         w.BiteService.Options,
			TargetNamespace: w.BiteService.TargetNamespace,
			Chart:           w.BiteService.Chart,
			Repo:            w.BiteService.Repo,
			Set:             w.BiteService.Set,
			ValuesContent:   w.BiteService.ValuesContent,
			Ignore:          w.BiteService.Ignore,
		},
	}

	return retval, nil
}

// ServiceMeshGateway extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) ServiceMeshGateway() (*ext.PrsnExternalResource, error) {
	retval := &ext.PrsnExternalResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
			Namespace: w.Namespace,
			Name:      w.BiteService.Name,
		},
		Spec: ext.PrsnExternalResourceSpec{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []*ext.Server{
				{
					Port: &ext.Port{
						Number:   80,
						Protocol: "HTTP",
						Name:     "http",
					},
					Hosts: []string{"*"},
				},
			},
		},
	}

	if w.BiteService.Ssl == "true" {
		tls := &ext.ServerTLSOptions{
			Mode:           "SIMPLE",
			CredentialName: w.BiteService.Name,
		}
		retval.Spec.Servers[0].TLS = tls
	} else {
		retval.Spec.Servers[0].TLS = nil
	}

	return retval, nil
}

// ServiceMeshVirtualService extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) ServiceMeshVirtualService() (*ext.PrsnExternalResource, error) {
	hosts := []string{}
	port := w.BiteService.Ports[0]

	for _, url := range w.BiteService.ExternalURL {
		hosts = append(hosts, url)
	}

	backend := w.BiteService.Name
	if w.BiteService.Backend != "" {
		backend = w.BiteService.Backend
	}

	retval := &ext.PrsnExternalResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
			Namespace: w.Namespace,
			Name:      w.BiteService.Name,
		},
		Spec: ext.PrsnExternalResourceSpec{
			Gateways: []string{
				w.BiteService.Name,
			},
			Hosts: hosts,
			HTTP: []*ext.HTTPRoute{
				{
					Match: []*ext.HTTPMatchRequest{
						{
							URI: &ext.StringExact{
								Exact: "/",
							},
						},
					},
					Route: []*ext.HTTPRouteDestination{
						{
							Destination: &ext.Destination{
								Host: backend,
								Port: &ext.PortSelector{
									Number: uint32(port),
								},
							},
						},
					},
				},
			},
		},
	}

	return retval, nil
}

func getAPIVersion(options map[string]interface{}) string {
	if options != nil && options["api_version"] != nil {
		return options["api_version"].(string)
	}
	return "prsn.io/v1"
}

func getAccessModesFromString(modes string) []v1.PersistentVolumeAccessMode {
	strmodes := strings.Split(modes, ",")
	var accessModes []v1.PersistentVolumeAccessMode
	for _, s := range strmodes {
		s = strings.Trim(s, " ")
		switch {
		case s == "ReadWriteOnce":
			accessModes = append(accessModes, v1.ReadWriteOnce)
		case s == "ReadOnlyMany":
			accessModes = append(accessModes, v1.ReadOnlyMany)
		case s == "ReadWriteMany":
			accessModes = append(accessModes, v1.ReadWriteMany)
		}
	}
	return accessModes
}

func (w *KubeMapper) resources() (v1.ResourceRequirements, error) {
	//Environment Operator allows for Guaranteed and Burstable QoS Classes as limits are always assigned to containers
	requests := v1.ResourceList{}
	limits := v1.ResourceList{}

	if quantity, err := resource.ParseQuantity(w.BiteService.Limits.CPU); err == nil {
		limits["cpu"] = quantity
	}

	if quantity, err := resource.ParseQuantity(w.BiteService.Limits.Memory); err == nil {
		limits["memory"] = quantity
	}

	if quantity, err := resource.ParseQuantity(w.BiteService.Requests.CPU); err == nil {
		requests["cpu"] = quantity
	}

	if quantity, err := resource.ParseQuantity(w.BiteService.Requests.Memory); err == nil {
		requests["memory"] = quantity
	}

	return v1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}, nil
}

func (w *KubeMapper) annotations() map[string]string {
	retval := map[string]string{}
	retval["deployment_method"] = w.BiteService.DeploymentMethod()
	if w.BiteService.IsBlueGreenParentDeployment() {
		retval["deployment_active"] = w.BiteService.ActiveDeploymentTag().String()
	}
	return retval
}

func (w *KubeMapper) labels() map[string]string {
	return map[string]string{
		"creator":     "pipeline",
		"application": w.BiteService.Application,
		"name":        w.BiteService.Name,
		"version":     w.BiteService.Version,
	}
}
