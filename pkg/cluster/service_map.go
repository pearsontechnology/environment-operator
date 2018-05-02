package cluster

import (
	"sort"
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	v1beta1_apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	autoscale_v1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	v1beta1_ext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// ServiceMap holds a list of bitesize.Service objects, representing the
// whole environment. Actions on it allow to fill in respective bits of
// information, from kubernetes objects to bitesize objects
type ServiceMap map[string]*bitesize.Service

func (s ServiceMap) CreateOrGet(name string) *bitesize.Service {
	// Create with some defaults -- defaults should probably live in bitesize.Service
	if s[name] == nil {
		s[name] = &bitesize.Service{
			Name:        name,
			Replicas:    1,
			Annotations: map[string]string{},
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

func (s ServiceMap) AddDeployment(deployment v1beta1_ext.Deployment) {
	name := deployment.Name

	biteservice := s.CreateOrGet(name)
	if deployment.Spec.Replicas != nil {
		biteservice.Replicas = int(*deployment.Spec.Replicas)
	}

	if len(deployment.Spec.Template.Spec.Containers[0].Resources.Requests) != 0 {
		cpuQuantity := new(resource.Quantity)
		*cpuQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"]
		memoryQuantity := new(resource.Quantity)
		*memoryQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
		biteservice.Requests.CPU = cpuQuantity.String()
		biteservice.Requests.Memory = memoryQuantity.String()
	}

	if len(deployment.Spec.Template.Spec.Containers[0].Resources.Limits) != 0 {
		cpuQuantity := new(resource.Quantity)
		*cpuQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"]
		memoryQuantity := new(resource.Quantity)
		*memoryQuantity = deployment.Spec.Template.Spec.Containers[0].Resources.Limits["memory"]
		biteservice.Limits.CPU = cpuQuantity.String()
		biteservice.Limits.Memory = memoryQuantity.String()
	}

	if getLabel(deployment.ObjectMeta, "ssl") != "" {
		biteservice.Ssl = getLabel(deployment.ObjectMeta, "ssl") // kubeDeployment.Labels["ssl"]
	}
	biteservice.Version = getLabel(deployment.ObjectMeta, "version")
	biteservice.Application = getLabel(deployment.ObjectMeta, "application")
	biteservice.HTTPSOnly = getLabel(deployment.ObjectMeta, "httpsOnly")
	biteservice.HTTPSBackend = getLabel(deployment.ObjectMeta, "httpsBackend")
	biteservice.EnvVars = envVars(deployment)
	biteservice.HealthCheck = healthCheck(deployment)

	for _, cmd := range deployment.Spec.Template.Spec.Containers[0].Command {
		biteservice.Commands = append(biteservice.Commands, string(cmd))
	}

	if deployment.Spec.Template.ObjectMeta.Annotations != nil {
		biteservice.Annotations = deployment.Spec.Template.ObjectMeta.Annotations
	} else {
		biteservice.Annotations = map[string]string{}
	}
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
			Path:  strings.Replace(claim.ObjectMeta.Labels["mount_path"], "2F", "/", -1),
			Modes: getAccessModesAsString(claim.Spec.AccessModes),
			Size:  claim.ObjectMeta.Labels["size"],
			Name:  claim.ObjectMeta.Name,
			Type:  claim.ObjectMeta.Labels["type"],
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

func (s ServiceMap) AddIngress(ingress v1beta1_ext.Ingress) {
	name := ingress.Name
	biteservice := s.CreateOrGet(name)
	ssl := ingress.Labels["ssl"]
	httpsOnly := ingress.Labels["httpsOnly"]
	httpsBackend := ingress.Labels["httpsBackend"]

	if len(ingress.Spec.Rules) > 0 {
		for _, rule := range ingress.Spec.Rules {
			biteservice.ExternalURL = append(biteservice.ExternalURL, rule.Host)
		}
	}

	biteservice.HTTPSBackend = httpsBackend
	biteservice.HTTPSOnly = httpsOnly
	biteservice.HTTP2 = ingress.Labels["http2"]
	biteservice.Ssl = ssl

	// backend service has been overridden
	backendService := ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName
	if backendService != biteservice.Name {
		biteservice.Backend = backendService
	}
	// backend port has been overriden
	backendPort := int(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal)
	if len(biteservice.Ports) > 0 && backendPort != biteservice.Ports[0] {
		biteservice.BackendPort = backendPort
	}
}

func (s ServiceMap) AddMongoStatefulSet(statefulset v1beta1_apps.StatefulSet) {
	name := statefulset.Name

	biteservice := s.CreateOrGet(name)

	if statefulset.Spec.Replicas != nil {
		biteservice.Replicas = int(*statefulset.Spec.Replicas)
	}

	if len(statefulset.Spec.Template.Spec.Containers[0].Resources.Requests) != 0 {
		cpuQuantity := new(resource.Quantity)
		*cpuQuantity = statefulset.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"]
		memoryQuantity := new(resource.Quantity)
		*memoryQuantity = statefulset.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
		biteservice.Requests.CPU = cpuQuantity.String()
		biteservice.Requests.Memory = memoryQuantity.String()
	}

	if len(statefulset.Spec.Template.Spec.Containers[0].Resources.Limits) != 0 {
		cpuQuantity := new(resource.Quantity)
		*cpuQuantity = statefulset.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"]
		memoryQuantity := new(resource.Quantity)
		*memoryQuantity = statefulset.Spec.Template.Spec.Containers[0].Resources.Limits["memory"]
		biteservice.Limits.CPU = cpuQuantity.String()
		biteservice.Limits.Memory = memoryQuantity.String()

	}

	biteservice.DatabaseType = "mongo"

	if getLabel(statefulset.ObjectMeta, "ssl") != "" {
		biteservice.Ssl = getLabel(statefulset.ObjectMeta, "ssl")
	}
	biteservice.Version = getLabel(statefulset.ObjectMeta, "version")
	biteservice.Application = getLabel(statefulset.ObjectMeta, "application")
	biteservice.HTTPSOnly = getLabel(statefulset.ObjectMeta, "httpsOnly")
	biteservice.HTTPSBackend = getLabel(statefulset.ObjectMeta, "httpsBackend")
	biteservice.HealthCheck = healthCheckStatefulset(statefulset)

	//Commands and Termination Period for mongo containers are hardcoded in the spec, so no need to sync up the Bitesize service

	//for _, cmd := range statefulset.Spec.Template.Spec.Containers[0].Command {
	//	biteservice.Commands = append(biteservice.Commands, string(cmd))
	//}
	//if statefulset.Spec.Template.Spec.TerminationGracePeriodSeconds != nil {
	//	biteservice.GracePeriod = statefulset.Spec.Template.Spec.TerminationGracePeriodSeconds
	//}

	if statefulset.Spec.Template.ObjectMeta.Annotations != nil {
		biteservice.Annotations = statefulset.Spec.Template.ObjectMeta.Annotations
	} else {
		biteservice.Annotations = map[string]string{}
	}

	for _, claim := range statefulset.Spec.VolumeClaimTemplates {
		vol := bitesize.Volume{
			Path:  claim.ObjectMeta.Labels["mount_path"],
			Modes: getAccessModesAsString(claim.Spec.AccessModes),
			Name:  claim.ObjectMeta.Name,
			Size:  claim.ObjectMeta.Labels["size"],
			Type:  claim.ObjectMeta.Labels["type"],
		}
		biteservice.Volumes = append(biteservice.Volumes, vol)
	}

	biteservice.Status = bitesize.ServiceStatus{

		CurrentReplicas: int(statefulset.Status.Replicas),
		DeployedAt:      statefulset.CreationTimestamp.String(),
	}
}
