package translator

// translator package converts objects between Kubernetes and Bitesize

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
	apps_v1 "k8s.io/api/apps/v1"
	autoscale_v2beta2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	netwk_v1beta1 "k8s.io/api/networking/v1beta1"
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
			Name:       fmt.Sprintf("http-%d", p),
		}

		if strings.EqualFold(w.BiteService.Protocol, "tcp") {
			servicePort.Name = fmt.Sprintf("tcp-port-%d", p)
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
			Name:       fmt.Sprintf("http-%d", p),
		}

		if strings.EqualFold(w.BiteService.Protocol, "tcp") {
			servicePort.Name = fmt.Sprintf("tcp-port-%d", p)
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
func (w *KubeMapper) Deployment() (*apps_v1.Deployment, error) {
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

	retval := &apps_v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
				"version":     w.BiteService.Version,
				"app":         w.BiteService.Application,
			},
		},
		Spec: apps_v1.DeploymentSpec{
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
						"app":         w.BiteService.Application,
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
func (w *KubeMapper) HPA() (*autoscale_v2beta2.HorizontalPodAutoscaler, error) {
	if w.BiteService.IsBlueGreenParentDeployment() {
		return nil, nil
	}
	retval := &autoscale_v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels:    w.labels(),
		},
		Spec: autoscale_v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v2beta2.CrossVersionObjectReference{
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

func (w *KubeMapper) getMetricSpec() (m []autoscale_v2beta2.MetricSpec) {
	if w.BiteService.HPA.Metric.Name == "cpu" || w.BiteService.HPA.Metric.Name == "memory" {
		if w.BiteService.HPA.Metric.Name == "cpu" && w.BiteService.HPA.Metric.TargetAverageUtilization != 0 {
			m = append(m, autoscale_v2beta2.MetricSpec{
				Type: autoscale_v2beta2.ResourceMetricSourceType,
				Resource: &autoscale_v2beta2.ResourceMetricSource{
					Target: autoscale_v2beta2.MetricTarget{
						Type:               "Utilization",
						AverageUtilization: &w.BiteService.HPA.Metric.TargetAverageUtilization,
					},
					Name: "cpu",
				},
			},
			)
		}
		if w.BiteService.HPA.Metric.Name == "memory" && w.BiteService.HPA.Metric.TargetAverageUtilization != 0 {
			m = append(m, autoscale_v2beta2.MetricSpec{
				Type: autoscale_v2beta2.ResourceMetricSourceType,
				Resource: &autoscale_v2beta2.ResourceMetricSource{
					Target: autoscale_v2beta2.MetricTarget{
						Type:               "Utilization",
						AverageUtilization: &w.BiteService.HPA.Metric.TargetAverageUtilization,
					},
					Name: "memory",
				},
			},
			)
		}
	} else {
		targetValue, _ := resource.ParseQuantity(w.BiteService.HPA.Metric.TargetAverageValue)
		m = append(m, autoscale_v2beta2.MetricSpec{
			Type: autoscale_v2beta2.PodsMetricSourceType,
			Pods: &autoscale_v2beta2.PodsMetricSource{
				Target: autoscale_v2beta2.MetricTarget{
					Type:         "AverageValue",
					AverageValue: &targetValue,
				},
				Metric: autoscale_v2beta2.MetricIdentifier{
					Name: w.BiteService.HPA.Metric.Name,
				},
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

	var ports []v1.ContainerPort
	for _, port := range w.BiteService.Ports {
		containerPort := v1.ContainerPort{
			ContainerPort: int32(port),
			Protocol:      "TCP",
		}
		ports = append(ports, containerPort)
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
		Ports:          ports,
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
func (w *KubeMapper) Ingress() (*netwk_v1beta1.Ingress, error) {
	labels := map[string]string{
		"creator":     "pipeline",
		"application": w.BiteService.Application,
		"name":        w.BiteService.Name,
	}

	labels["ssl"] = w.BiteService.Ssl

	if w.BiteService.HTTPSBackend != "" {
		labels["httpsBackend"] = w.BiteService.HTTPSBackend
	}

	labels["httpsOnly"] = w.BiteService.HTTPSOnly

	if w.BiteService.HTTP2 != "" {
		labels["http2"] = w.BiteService.HTTP2
	}

	port := intstr.FromInt(w.BiteService.Ports[0])
	retval := &netwk_v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Spec: netwk_v1beta1.IngressSpec{
			Rules: []netwk_v1beta1.IngressRule{},
		},
	}

	for _, url := range w.BiteService.ExternalURL {
		rule := netwk_v1beta1.IngressRule{
			Host: url,
			IngressRuleValue: netwk_v1beta1.IngressRuleValue{
				HTTP: &netwk_v1beta1.HTTPIngressRuleValue{
					Paths: []netwk_v1beta1.HTTPIngressPath{
						{
							Path: "/",
							Backend: netwk_v1beta1.IngressBackend{
								ServiceName: w.BiteService.Name,
								ServicePort: port,
							},
						},
					},
				},
			},
		}

		if w.BiteService.Ssl == "true" {
			tls := netwk_v1beta1.IngressTLS{
				Hosts:      []string{url},
				SecretName: w.BiteService.Name,
			}
			retval.Spec.TLS = []netwk_v1beta1.IngressTLS{tls}
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

func (w *KubeMapper) ExternalSecretTLS() (*ext.ExternalSecret, error) {

	labels := map[string]string{
		"creator":     "pipeline",
		"application": w.BiteService.Application,
		"name":        w.BiteService.Name,
	}

	var env, envType string

	if env = os.Getenv("ENVIRONMENT"); env == "" {
		return nil, errors.New("ENVIRONMENT is not set")
	}

	if envType = os.Getenv("ENVTYPE"); envType == "" {
		return nil, errors.New("ENVTYPE is not set")
	}

	return &ext.ExternalSecret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes-client.io/v1",
			Kind:       strings.Title("ExternalSecret"),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   w.BiteService.Name,
			Labels: labels,
		},
		SecretDescriptor: ext.ExternalSecretSecretDescriptor{
			BackendType: "secretsManager",
			Compressed:  true,
			Type:        "kubernetes.io/tls",
			Data: []map[string]string{
				{
					"key":  fmt.Sprintf("tls/%s/%s/%s/%s.crt", envType, env, w.Namespace, w.BiteService.Name),
					"name": "tls.crt",
				},
				{
					"key":  fmt.Sprintf("tls/%s/%s/%s/%s.key", envType, env, w.Namespace, w.BiteService.Name),
					"name": "tls.key",
				},
			},
		},
	}, nil
}

// CustomResourceDefinition extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) CustomResourceDefinition() (*ext.PrsnExternalResource, error) {
	ports := []*ext.Port{}
	for _, port := range w.BiteService.ServiceEntryPorts {
		ports = append(ports, &ext.Port{
			Number:   port.Number,
			Protocol: port.Protocol,
			Name:     port.Name,
		})
	}

	endpoints := []*ext.ServiceEntry_Endpoint{}
	for _, endpoint := range w.BiteService.Endpoints {
		endpoints = append(endpoints, &ext.ServiceEntry_Endpoint{
			Address:  endpoint.Address,
			Ports:    endpoint.Ports,
			Labels:   endpoint.Labels,
			Network:  endpoint.Network,
			Locality: endpoint.Locality,
			Weight:   endpoint.Weight,
		})
	}

	retval := &ext.PrsnExternalResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       strings.Title(w.BiteService.Type),
			APIVersion: getAPIVersion(w.BiteService.Type),
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
			Hosts:           w.BiteService.Hosts,
			Addresses:       w.BiteService.Addresses,
			Ports:           ports,
			Location:        w.BiteService.Location,
			Resolution:      w.BiteService.Resolution,
			Endpoints:       endpoints,
		},
	}

	return retval, nil
}

// ServiceMeshGateway extracts Kubernetes object from BiteSize definition
func (w *KubeMapper) ServiceMeshGateway() (*ext.PrsnExternalResource, error) {
	hosts := []string{}
	for _, url := range w.BiteService.ExternalURL {
		hosts = append(hosts, url)
	}

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
					Hosts: hosts,
				},
			},
		},
	}

	if w.BiteService.Ssl == "true" {
		servers := []*ext.Server{
			{
				Port: &ext.Port{
					Number:   443,
					Protocol: "HTTPS",
					Name:     "https",
				},
				Hosts: hosts,
			},
			{
				Port: &ext.Port{
					Number:   80,
					Protocol: "HTTP",
					Name:     "http",
				},
				Hosts: hosts,
			},
		}
		retval.Spec.Servers = servers

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
							URI: &ext.StringPrefix{
								Prefix: "/",
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

func getAPIVersion(Type string) string {
	switch strings.ToLower(Type) {
	case "helmchart":
		return "helm.kubedex.com/v1"
	case "serviceentry":
		return "networking.istio.io/v1alpha3"
	default:
		return "prsn.io/v1"
	}
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
