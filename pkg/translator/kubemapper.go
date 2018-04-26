package translator

// translator package converts objects between Kubernetes and Bitesize

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/base64"
	"math/rand"
	"time"

	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	v1beta1_apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	autoscale_v1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	v1beta1_ext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

// KubeMapper maps BitesizeService object to Kubernetes objects
type KubeMapper struct {
	BiteService *bitesize.Service
	Namespace   string
	Config      struct {
		Project        string
		DockerRegistry string
	}
}

// Service extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) Service() (*v1.Service, error) {
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
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
		},
	}
	return retval, nil
}

// Service extracts Kubernetes Headless Service object (No ClusterIP) from Bitesize definition
func (w *KubeMapper) HeadlessService() (*v1.Service, error) {
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
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
			ClusterIP: v1.ClusterIPNone,
		},
	}
	return retval, nil
}

// PersistentVolumeClaims returns a list of claims for a biteservice
func (w *KubeMapper) PersistentVolumeClaims() ([]v1.PersistentVolumeClaim, error) {
	var retval []v1.PersistentVolumeClaim

	for _, vol := range w.BiteService.Volumes {
		ret := v1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      vol.Name,
				Namespace: w.Namespace,
				Labels: map[string]string{
					"creator":    "pipeline",
					"deployment": w.BiteService.Name,
					"mount_path": strings.Replace(vol.Path, "/", "2F", -1),
					"size":       vol.Size,
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
			ret.Spec.Selector = &unversioned.LabelSelector{
				MatchLabels: map[string]string{
					"name": vol.Name,
				},
			}
		}

		retval = append(retval, ret)
	}
	return retval, nil
}

// MongoInternalSecret returns a secret to be used for mongo internal Auth
func (w *KubeMapper) MongoInternalSecret() (*v1.Secret, error) {

	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 700)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	value := base64.StdEncoding.EncodeToString(b)

	s := map[string]string{
		"internal-auth-mongodb-keyfile": value,
	}

	ret := &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "mongo-bootstrap-data",
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":    "pipeline",
				"deployment": w.BiteService.Name,
			},
		},
		StringData: s,
	}

	return ret, nil
}

// Stateful set extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) MongoStatefulSet() (*v1beta1_apps.StatefulSet, error) {
	replicas := int32(w.BiteService.Replicas)
	imagePullSecrets, err := w.imagePullSecrets()
	if err != nil {
		return nil, err
	}
	mounts, err := w.volumeMounts()
	if err != nil {
		return nil, err
	}
	vol := v1.VolumeMount{
		Name:      "secrets-volume",
		MountPath: "/etc/secrets-volume",
		ReadOnly:  true,
	}

	mounts = append(mounts, vol)

	resources, err := w.resources()
	if err != nil {
		return nil, err
	}

	retval := &v1beta1_apps.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
				"version":     w.BiteService.Version,
			},
		},
		Spec: v1beta1_apps.StatefulSetSpec{
			ServiceName: w.BiteService.Name,
			Replicas:    &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Name:      w.BiteService.Name,
					Namespace: w.Namespace,
					Labels: map[string]string{
						"creator":     "pipeline",
						"application": w.BiteService.Application,
						"name":        w.BiteService.Name,
						"version":     w.BiteService.Version,
						"role":        "mongo",
					},
					Annotations: w.BiteService.Annotations,
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &[]int64{10}[0],
					Volumes: []v1.Volume{
						{
							Name: "secrets-volume",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName:  "mongo-bootstrap-data",
									DefaultMode: &[]int32{256}[0],
								},
							},
						},
					},
					NodeSelector: map[string]string{"role": "minion"},
					Containers: []v1.Container{
						{
							Name:            w.BiteService.DatabaseType,
							Image:           fmt.Sprintf("%s:%s", w.BiteService.DatabaseType, w.BiteService.Version),
							ImagePullPolicy: v1.PullAlways,
							Command: []string{
								"mongod",
								"--replSet",
								"mongo",
								"--auth",
								"--smallfiles",
								"--noprealloc",
								"--clusterAuthMode",
								"keyFile",
								"--keyFile",
								"/etc/secrets-volume/internal-auth-mongodb-keyfile",
								"--setParameter",
								"authenticationMechanisms=SCRAM-SHA-1",
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: int32(w.BiteService.Ports[0]),
								},
							},
							VolumeMounts: mounts,
							Resources:    resources,
						},
					},
					ImagePullSecrets: imagePullSecrets,
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      w.BiteService.Volumes[0].Name,
						Namespace: w.Namespace,
						Annotations: map[string]string{
							"volume.beta.kubernetes.io/storage-class": "aws-ebs",
						},
						Labels: map[string]string{
							"creator":    "pipeline",
							"deployment": w.BiteService.Name,
							"mount_path": w.BiteService.Volumes[0].Path,
							"size":       w.BiteService.Volumes[0].Size,
						},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: getAccessModesFromString(w.BiteService.Volumes[0].Modes),
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceName(v1.ResourceStorage): resource.MustParse(w.BiteService.Volumes[0].Size),
							},
						},
					},
				},
			},
		},
	}
	return retval, nil
}

// Deployment extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) Deployment() (*v1beta1_ext.Deployment, error) {
	replicas := int32(w.BiteService.Replicas)
	container, err := w.container()
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
		ObjectMeta: v1.ObjectMeta{
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
			Selector: &unversioned.LabelSelector{
				MatchLabels: map[string]string{
					"creator": "pipeline",
					"name":    w.BiteService.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
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
func (w *KubeMapper) HPA() (autoscale_v1.HorizontalPodAutoscaler, error) {
	retval := autoscale_v1.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
				"version":     w.BiteService.Version,
			},
		},
		Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       w.BiteService.Name,
				APIVersion: "v1beta1",
			},
			MinReplicas:                    &w.BiteService.HPA.MinReplicas,
			MaxReplicas:                    w.BiteService.HPA.MaxReplicas,
			TargetCPUUtilizationPercentage: &w.BiteService.HPA.TargetCPUUtilizationPercentage,
		},
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
	retval = &v1.Container{
		Name:         w.BiteService.Name,
		Image:        "",
		Env:          evars,
		VolumeMounts: mounts,
		Resources:    resources,
		Command:      w.BiteService.Commands,
	}

	return retval, nil
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
				log.Debugf("Unable to Find Secret %s", secretName)
				err = fmt.Errorf("Unable to find secret [%s] in namespace [%s] when processing envvars for deployment [%s]", secretName, config.Env.Namespace, w.BiteService.Name)
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
	return retval, err
}

func (w *KubeMapper) volumeMounts() ([]v1.VolumeMount, error) {
	var retval []v1.VolumeMount

	for _, v := range w.BiteService.Volumes {
		if v.Name == "" || v.Path == "" {
			return nil, fmt.Errorf("Volume must have both name and path set")
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
			Name: v.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.Name,
				},
			},
		}
		retval = append(retval, vol)
	}
	return retval, nil
}

// Ingress extracts Kubernetes object from Bitesize definition
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
		ObjectMeta: v1.ObjectMeta{
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

// ThirdPartyResource extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) ThirdPartyResource() (*ext.PrsnExternalResource, error) {
	retval := &ext.PrsnExternalResource{
		TypeMeta: unversioned.TypeMeta{
			Kind:       strings.Title(w.BiteService.Type),
			APIVersion: "prsn.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"creator": "pipeline",
				"name":    w.BiteService.Name,
			},
			Namespace: w.Namespace,
			Name:      w.BiteService.Name,
		},
		Spec: ext.PrsnExternalResourceSpec{
			Version: w.BiteService.Version,
			Options: w.BiteService.Options,
		},
	}

	log.Debugf("PrsnExternalResource: %+v", *retval)

	return retval, nil
}

func getAccessModesFromString(modes string) []v1.PersistentVolumeAccessMode {
	strmodes := strings.Split(modes, ",")
	accessModes := []v1.PersistentVolumeAccessMode{}
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
	cpuRequest, memoryRequestError := resource.ParseQuantity(w.BiteService.Requests.CPU)
	memoryRequest, cpuRequestError := resource.ParseQuantity(w.BiteService.Requests.Memory)
	cpuLimit, _ := resource.ParseQuantity(w.BiteService.Limits.CPU)
	memoryLimit, _ := resource.ParseQuantity(w.BiteService.Limits.Memory)

	if cpuRequestError != nil && memoryRequestError != nil { //If no CPU or Memory Request provided, default to limits for Guaranteed QoS
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
		}, nil
	} else if cpuRequestError != nil && memoryRequestError == nil {
		return v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": memoryRequest,
			},
			Limits: v1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
		}, nil
	} else if cpuRequestError == nil && memoryRequestError != nil {
		return v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"cpu": cpuRequest,
			},
			Limits: v1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
		}, nil
	} else {
		return v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"cpu":    cpuRequest,
				"memory": memoryRequest,
			},
			Limits: v1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
		}, nil

	}
}
