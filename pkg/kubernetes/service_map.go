package kubernetes

import (
	"sort"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
)

// ServiceMap holds a list of bitesize.Service objects, representing the
// whole environment. Actions on it allow to fill in respective bits of
// information, from kubernetes objects to bitesize objects
type ServiceMap map[string]*bitesize.Service

// Services extracts a sorted list of bitesize.Services type out from
// ServiceMap type
func (s ServiceMap) Services() bitesize.Services {
	var serviceList bitesize.Services

	for _, v := range s {
		serviceList = append(serviceList, *v)
	}

	sort.Sort(serviceList)
	return serviceList
}

func (s ServiceMap) AddService(svc v1.Service) {
	name := svc.Name
	if s[name] == nil {
		s[name] = &bitesize.Service{Name: name}
	}

	if len(svc.Spec.Ports) > 0 {
		s[name].Port = int(svc.Spec.Ports[0].Port)
	} else {
		s[name].Port = 0
	}

	if s[name].Replicas == 0 {
		s[name].Replicas = 1
	}
}

func (s ServiceMap) AddDeployment(deployment v1beta1.Deployment) {
	name := deployment.Name

	if s[name] == nil {
		s[name] = &bitesize.Service{Name: name}
	}

	if deployment.Spec.Replicas != nil {
		s[name].Replicas = int(*deployment.Spec.Replicas)
	} else {
		s[name].Replicas = 1
	}
	s[name].Ssl = getLabel(deployment, "ssl") // kubeDeployment.Labels["ssl"]
	s[name].Version = getLabel(deployment, "version")
	s[name].Application = getLabel(deployment, "application")
	s[name].HTTPSOnly = getLabel(deployment, "httpsOnly")
	s[name].HTTPSBackend = getLabel(deployment, "httpsBackend")
	s[name].EnvVars = envVars(deployment)
	s[name].HealthCheck = healthCheck(deployment)
}

func (s ServiceMap) AddVolumeClaim(claim v1.PersistentVolumeClaim) {
	name := claim.ObjectMeta.Labels["deployment"]

	if name != "" {
		if s[name] == nil {
			s[name] = &bitesize.Service{Name: name}
		}
		vol := bitesize.Volume{
			Path:  claim.ObjectMeta.Labels["mount_path"],
			Modes: getAccessModesAsString(claim.Spec.AccessModes),
			Size:  claim.ObjectMeta.Labels["size"],
			Name:  claim.ObjectMeta.Name,
		}
		s[name].Volumes = append(s[name].Volumes, vol)
	}
}

func (s ServiceMap) AddIngress(ingress v1beta1.Ingress) {
	var externalURL string

	name := ingress.Name

	if s[name] == nil {
		s[name] = &bitesize.Service{Name: name}
	}

	if len(ingress.Spec.Rules) > 0 {
		externalURL = ingress.Spec.Rules[0].Host
	}

	s[name].ExternalURL = externalURL
}
