package translator

// translator package converts objects between Kubernetes and Bitesize

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	"k8s.io/client-go/pkg/api/unversioned"
	api_v1 "k8s.io/client-go/pkg/api/v1"
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
func (w *KubeMapper) Service() (*api_v1.Service, error) {
	retval := &api_v1.Service{
		ObjectMeta: api_v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name,
				"application": w.BiteService.Application,
			},
		},
		Spec: api_v1.ServiceSpec{
			Ports: []api_v1.ServicePort{
				{
					Port:       int32(w.BiteService.Port),
					TargetPort: intstr.FromInt(w.BiteService.Port),
					Name:       "tcp-port",
				},
			},
		},
	}
	return retval, nil
}

// PersistentVolumeClaims returns a list of claims for a biteservice
func (w *KubeMapper) PersistentVolumeClaims() ([]api_v1.PersistentVolumeClaim, error) {
	var retval []api_v1.PersistentVolumeClaim

	for _, vol := range w.BiteService.Volumes {
		ret := api_v1.PersistentVolumeClaim{
			ObjectMeta: api_v1.ObjectMeta{
				Name:      vol.Name,
				Namespace: w.Namespace,
				Labels: map[string]string{
					"creator":    "pipeline",
					"deployment": w.BiteService.Name,
					"mount_path": vol.Path,
					"size":       vol.Size,
				},
			},
			Spec: api_v1.PersistentVolumeClaimSpec{
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
		container.Image, _ = util.ApplicationImage(w.BiteService)
	}

	volumes, err := w.volumes()
	if err != nil {
		return nil, err
	}

	retval := &v1beta1.Deployment{
		ObjectMeta: api_v1.ObjectMeta{
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
			Template: api_v1.PodTemplateSpec{
				ObjectMeta: api_v1.ObjectMeta{
					Name:      w.BiteService.Name,
					Namespace: w.Namespace,
					Labels: map[string]string{
						"creator":     "pipeline",
						"application": w.BiteService.Application,
						"name":        w.BiteService.Name,
						"version":     w.BiteService.Version,
					},
				},
				Spec: api_v1.PodSpec{
					NodeSelector: map[string]string{"role": "minion"},
					Containers:   []api_v1.Container{*container},
					Volumes:      volumes,
				},
			},
		},
	}

	return retval, nil
}

func (w *KubeMapper) container() (*api_v1.Container, error) {
	mounts, err := w.volumeMounts()
	if err != nil {
		return nil, err
	}

	evars, err := w.envVars()
	if err != nil {
		return nil, err
	}

	retval := &api_v1.Container{
		Name:         w.BiteService.Name,
		Image:        "",
		Env:          evars,
		VolumeMounts: mounts,
	}
	return retval, nil
}

func (w *KubeMapper) envVars() ([]api_v1.EnvVar, error) {
	var retval []api_v1.EnvVar

	for _, e := range w.BiteService.EnvVars {
		evar := api_v1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		}
		retval = append(retval, evar)
	}
	return retval, nil
}

func (w *KubeMapper) volumeMounts() ([]api_v1.VolumeMount, error) {
	var retval []api_v1.VolumeMount

	for _, v := range w.BiteService.Volumes {
		if v.Name == "" || v.Path == "" {
			return nil, fmt.Errorf("Volume must have both name and path set")
		}
		vol := api_v1.VolumeMount{
			Name:      v.Name,
			MountPath: v.Path,
		}
		retval = append(retval, vol)
	}
	return retval, nil
}

func (w *KubeMapper) volumes() ([]api_v1.Volume, error) {
	var retval []api_v1.Volume
	for _, v := range w.BiteService.Volumes {
		vol := api_v1.Volume{
			Name: v.Name,
			VolumeSource: api_v1.VolumeSource{
				PersistentVolumeClaim: &api_v1.PersistentVolumeClaimVolumeSource{
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
	port := intstr.FromInt(w.BiteService.Port)
	retval := &v1beta1.Ingress{
		ObjectMeta: api_v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"application": w.BiteService.Application,
				"name":        w.BiteService.Name,
			},
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
		ObjectMeta: api_v1.ObjectMeta{
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

func getAccessModesFromString(modes string) []api_v1.PersistentVolumeAccessMode {
	strmodes := strings.Split(modes, ",")
	accessModes := []api_v1.PersistentVolumeAccessMode{}
	for _, s := range strmodes {
		s = strings.Trim(s, " ")
		switch {
		case s == "ReadWriteOnce":
			accessModes = append(accessModes, api_v1.ReadWriteOnce)
		case s == "ReadOnlyMany":
			accessModes = append(accessModes, api_v1.ReadOnlyMany)
		case s == "ReadWriteMany":
			accessModes = append(accessModes, api_v1.ReadWriteMany)
		}
	}
	return accessModes
}
