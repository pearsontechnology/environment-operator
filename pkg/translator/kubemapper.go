package translator

// translator package converts objects between Kubernetes and Bitesize

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	autoscale_v1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
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
					"mount_path": vol.Path,
					"size":       vol.Size,
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				VolumeName:  vol.Name,
				AccessModes: getAccessModesFromString(vol.Modes),
				Selector: &unversioned.LabelSelector{
					MatchLabels: map[string]string{
						"name": vol.Name,
					},
				},
			},
		}

		retval = append(retval, ret)
	}
	return retval, nil
}

// Deployment extracts Kubernetes object from Bitesize definition
func (w *KubeMapper) Deployment() (*v1beta1.Deployment, error) {
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

	retval := &v1beta1.Deployment{
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
		Spec: v1beta1.DeploymentSpec{
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

	result := strings.Split(util.RegistrySecrets(), ",")
	for i := range result {
		var namevalue v1.LocalObjectReference
		namevalue = v1.LocalObjectReference{
			Name: result[i],
		}
		retval = append(retval, namevalue)
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

	annotations, err := w.Annotations()
	if err != nil {
		return nil, err
	}

	if w.BiteService.Requests != (bitesize.ContainerRequests{}) {
		resources, err := w.resources()
		if err != nil {
			return nil, err
		}
		retval = &v1.Container{
			Name:         w.BiteService.Name,
			Image:        "",
			Env:          evars,
			Annotation:   annotations,
			VolumeMounts: mounts,
			Resources:    resources,
		}
	} else {
		retval = &v1.Container{
			Name:         w.BiteService.Name,
			Image:        "",
			Env:          evars,
			Annotation:   annotations,
			VolumeMounts: mounts,
		}
	}

	return retval, nil
}

func (w *KubeMapper) envVars() ([]v1.EnvVar, error) {
	var retval []v1.EnvVar

	for _, e := range w.BiteService.EnvVars {
		var evar v1.EnvVar
		if e.Secret != "" {
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
		} else {
			evar = v1.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			}
		}
		retval = append(retval, evar)
	}
	return retval, nil
}

func (w *KubeMapper) Annotations() ([]v1.Annotation, error) {
	var retval []v1.Annotation

	for _, e := range w.BiteService.Annotation {
		var annotation v1.Annotation
		evar = v1.Annotation{
			Name:  e.Name,
			Value: e.Value,
		retval = append(retval, evar)
	}
	return retval, nil
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
func (w *KubeMapper) Ingress() (*v1beta1.Ingress, error) {
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

	port := intstr.FromInt(w.BiteService.Ports[0])
	retval := &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: w.BiteService.ExternalURL,

					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: w.BiteService.Name,
										ServicePort: port,
									},
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
	cpuQuantity, err := resource.ParseQuantity(w.BiteService.Requests.CPU)
	if err != nil {
		return v1.ResourceRequirements{}, err
	}
	//	memoryQuantity, err := resource.ParseQuantity(w.BiteService.Requests.Memory)
	//	if err != nil {
	//		return v1.ResourceRequirements{}, err
	//	}
	retval := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			"cpu": cpuQuantity,
			//			"memory": memoryQuantity,
		},
	}
	return retval, nil
}
