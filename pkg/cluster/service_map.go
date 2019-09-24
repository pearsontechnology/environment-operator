package cluster

import (
	"sort"
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	autoscale_v2beta1 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	v1beta1_ext "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceMap holds a list of bitesize.Service objects, representing the
// whole environment. Actions on it allow to fill in respective bits of
// information, from kubernetes objects to bitesize objects
type ServiceMap map[string]*bitesize.Service

// CreateOrGet initializes new biteservice or returns an existing one (by name)
func (s ServiceMap) CreateOrGet(name string) *bitesize.Service {
	// Create with some defaults -- defaults should probably live in bitesize.Service
	if s[name] == nil {
		s[name] = &bitesize.Service{
			Name:        name,
			Replicas:    1,
			Annotations: map[string]string{},
			ExternalURL: []string{},
			Deployment:  &bitesize.DeploymentSettings{},
			Requests:    bitesize.ContainerRequests{CPU: "100m"},
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

// AddService adds Kubernetes service object to biteservice
func (s ServiceMap) AddService(svc v1.Service) {
	name := svc.Name
	biteservice := s.CreateOrGet(name)
	biteservice.Application = getLabel(svc.ObjectMeta, "application")
	biteservice.Deployment = s.addDeploymentSettings(svc.ObjectMeta)

	if len(svc.Spec.Ports) > 0 {
		biteservice.Ports = []int{}
	}
	for _, port := range svc.Spec.Ports {
		biteservice.Ports = append(biteservice.Ports, int(port.Port))
	}
}

func (s ServiceMap) addDeploymentSettings(metadata metav1.ObjectMeta) *bitesize.DeploymentSettings {
	retval := &bitesize.DeploymentSettings{}
	retval.Method = getAnnotation(metadata, "deployment_method")
	active := getAnnotation(metadata, "deployment_active")
	if active != "" {
		id := bitesize.BlueGreenDeploymentID(active)
		if id != 0 {
			retval.BlueGreen = &bitesize.BlueGreenSettings{Active: &id}
		}
	}
	return retval
}

// AddDeployment adds kubernetes deployment object to biteservice
func (s ServiceMap) AddDeployment(deployment v1beta1_ext.Deployment) {
	name := deployment.Name

	biteservice := s.CreateOrGet(name)
	if deployment.Spec.Replicas != nil {
		biteservice.Replicas = int(*deployment.Spec.Replicas)
	}

	resources := deployment.Spec.Template.Spec.Containers[0].Resources

	if len(resources.Requests) != 0 {
		cpuQuantity := resources.Requests["cpu"]
		memQuantity := resources.Requests["memory"]
		biteservice.Requests.CPU = cpuQuantity.String()
		biteservice.Requests.Memory = memQuantity.String()
	}

	if len(resources.Limits) != 0 {
		cpuQuantity := resources.Limits["cpu"]
		memQuantity := resources.Limits["memory"]
		biteservice.Limits.CPU = cpuQuantity.String()
		biteservice.Limits.Memory = memQuantity.String()
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
	biteservice.LivenessProbe = livenessProbe(deployment)
	biteservice.ReadinessProbe = readinessProbe(deployment)
	biteservice.Volumes = append(biteservice.Volumes, volumes(deployment)...)

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

// AddHPA adds Kubernetes HPA to biteservice
func (s ServiceMap) AddHPA(hpa autoscale_v2beta1.HorizontalPodAutoscaler) {
	name := hpa.Name

	biteservice := s.CreateOrGet(name)

	biteservice.HPA.MinReplicas = *hpa.Spec.MinReplicas
	biteservice.HPA.MaxReplicas = hpa.Spec.MaxReplicas

	if hpa.Spec.Metrics[0].Type == "Resource" {
		if hpa.Spec.Metrics[0].Resource.Name == "cpu" {
			biteservice.HPA.Metric.Name = "cpu"
		}

		if hpa.Spec.Metrics[0].Resource.Name == "memory" {
			biteservice.HPA.Metric.Name = "memory"
		}

		biteservice.HPA.Metric.TargetAverageUtilization = *hpa.Spec.Metrics[0].Resource.TargetAverageUtilization
	}

	if hpa.Spec.Metrics[0].Type == "Pods" {
		targetAverageValueQuantity := hpa.Spec.Metrics[0].Pods.TargetAverageValue
		biteservice.HPA.Metric.Name = hpa.Spec.Metrics[0].Pods.MetricName
		biteservice.HPA.Metric.TargetAverageValue = targetAverageValueQuantity.String()
	}

}

// AddVolumeClaim adds Kubernetes PVC to biteservice
func (s ServiceMap) AddVolumeClaim(claim v1.PersistentVolumeClaim) {
	name := claim.ObjectMeta.Labels["deployment"]

	if name == "" {
		return
	}

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

// AddCustomResourceDefinition adds Kubernetes CRD to biteservice
func (s ServiceMap) AddCustomResourceDefinition(crd k8_extensions.PrsnExternalResource) {
	name := crd.ObjectMeta.Name
	biteservice := s.CreateOrGet(name)
	biteservice.Type = strings.ToLower(crd.Kind)
	biteservice.Options = crd.Spec.Options
	biteservice.Version = crd.Spec.Version
	biteservice.TargetNamespace = crd.Spec.TargetNamespace
	biteservice.Chart = crd.Spec.Chart
	biteservice.Repo = crd.Spec.Repo
	biteservice.Set = crd.Spec.Set
	biteservice.ValuesContent = crd.Spec.ValuesContent
	biteservice.Ignore = crd.Spec.Ignore

	if crd.Spec.Replicas != 0 {
		biteservice.Replicas = crd.Spec.Replicas
	}
}

// AddIngress adds Kubernetes ingress fields to biteservice
func (s ServiceMap) AddIngress(ingress v1beta1_ext.Ingress) {
	name := ingress.Name
	biteservice := s.CreateOrGet(name)
	ssl := ingress.Labels["ssl"]
	httpsOnly := ingress.Labels["httpsOnly"]
	httpsBackend := ingress.Labels["httpsBackend"]

	biteservice.ExternalURL = []string{}
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
