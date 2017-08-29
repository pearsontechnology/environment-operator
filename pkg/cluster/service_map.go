package cluster

import (
	"sort"
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	autoscale_v1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// ServiceMap holds a list of bitesize.Service objects, representing the
// whole environment. Actions on it allow to fill in respective bits of
// information, from kubernetes objects to bitesize objects
type ServiceMap map[string]*bitesize.Service

func (s ServiceMap) CreateOrGet(name string) *bitesize.Service {
	// Create with some defaults -- defaults should probably live in bitesize.Service
	if s[name] == nil {
		s[name] = &bitesize.Service{
			Name:     name,
			Replicas: 1,
		}
	}
	return s[name]
}

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
	biteservice := s.CreateOrGet(name)
	biteservice.Application = getLabel(svc.ObjectMeta, "application")

	for _, port := range svc.Spec.Ports {
		biteservice.Ports = append(biteservice.Ports, int(port.Port))
	}
}

func (s ServiceMap) AddDeployment(deployment v1beta1.Deployment) {
	name := deployment.Name

	biteservice := s.CreateOrGet(name)
	if deployment.Spec.Replicas != nil {
		biteservice.Replicas = int(*deployment.Spec.Replicas)
	}

	if len(deployment.Spec.Template.Spec.Containers[0].Resources.Requests) != 0 {
		cpuQuantity := new(resource.Quantity)
		*cpuQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"]
		//		memoryQuantity := new(resource.Quantity)
		//		*memoryQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
		biteservice.Requests.CPU = cpuQuantity.String()
		//		biteservice.Requests.Memory = memoryQuantity.String()
	}

	biteservice.Ssl = getLabel(deployment.ObjectMeta, "ssl") // kubeDeployment.Labels["ssl"]
	biteservice.Version = getLabel(deployment.ObjectMeta, "version")
	biteservice.Application = getLabel(deployment.ObjectMeta, "application")
	biteservice.HTTPSOnly = getLabel(deployment.ObjectMeta, "httpsOnly")
	biteservice.HTTPSBackend = getLabel(deployment.ObjectMeta, "httpsBackend")
	biteservice.EnvVars = envVars(deployment)
	biteservice.Annotations = annVars(deployment.ObjectMeta)
	biteservice.HealthCheck = healthCheck(deployment)
	biteservice.Status = bitesize.ServiceStatus{

		AvailableReplicas: int(deployment.Status.AvailableReplicas),
		DesiredReplicas:   int(deployment.Status.Replicas),
		CurrentReplicas:   int(deployment.Status.UpdatedReplicas),
		DeployedAt:        deployment.CreationTimestamp.String(),
	}
}

func (s ServiceMap) AddHPA(hpa autoscale_v1.HorizontalPodAutoscaler) {
	name := hpa.Name

	biteservice := s.CreateOrGet(name)

	biteservice.HPA.MinReplicas = *hpa.Spec.MinReplicas
	biteservice.HPA.MaxReplicas = hpa.Spec.MaxReplicas
	biteservice.HPA.TargetCPUUtilizationPercentage = *hpa.Spec.TargetCPUUtilizationPercentage
}

func (s ServiceMap) AddVolumeClaim(claim v1.PersistentVolumeClaim) {
	name := claim.ObjectMeta.Labels["deployment"]

	if name != "" {
		biteservice := s.CreateOrGet(name)

		vol := bitesize.Volume{
			Path:  claim.ObjectMeta.Labels["mount_path"],
			Modes: getAccessModesAsString(claim.Spec.AccessModes),
			Size:  claim.ObjectMeta.Labels["size"],
			Name:  claim.ObjectMeta.Name,
		}
		biteservice.Volumes = append(biteservice.Volumes, vol)
	}
}

func (s ServiceMap) AddThirdPartyResource(tpr k8_extensions.PrsnExternalResource) {
	name := tpr.ObjectMeta.Name
	biteservice := s.CreateOrGet(name)
	biteservice.Type = strings.ToLower(tpr.Kind)
	biteservice.Options = tpr.Spec.Options
	biteservice.Version = tpr.Spec.Version
	if tpr.Spec.Replicas != 0 {
		biteservice.Replicas = tpr.Spec.Replicas
	}
}

func (s ServiceMap) AddPod(pod v1.Pod, logs string, error string) {
	biteservice := s.CreateOrGet("podservice")
	podval := bitesize.Pod{
		Name:      pod.ObjectMeta.Name,
		Phase:     pod.Status.Phase,
		StartTime: pod.Status.StartTime.String(),
		Message:   error,
		Logs:      logs,
	}
	biteservice.DeployedPods = append(biteservice.DeployedPods, podval)
}

func (s ServiceMap) AddIngress(ingress v1beta1.Ingress) {
	var externalURL string

	name := ingress.Name
	biteservice := s.CreateOrGet(name)

	ssl := ingress.Labels["ssl"]
	httpsOnly := ingress.Labels["httpsOnly"]
	httpsBackend := ingress.Labels["httpsBackend"]

	if len(ingress.Spec.Rules) > 0 {
		externalURL = ingress.Spec.Rules[0].Host
	}

	biteservice.ExternalURL = externalURL
	biteservice.HTTPSBackend = httpsBackend
	biteservice.HTTPSOnly = httpsOnly
	biteservice.Ssl = ssl
}
