package cluster

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client returns default in-cluster kubernetes client
func Client() (*Cluster, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	crdcli, err := k8s.CRDClient()
	if err != nil {
		return nil, err
	}

	return &Cluster{Interface: clientset, CRDClient: crdcli}, nil
}

// ApplyIfChanged compares bitesize Environment passed as an argument to
// the current client environment. If there are any changes, c is applied
// to the current config
func (cluster *Cluster) ApplyIfChanged(newConfig *bitesize.Environment) error {
	var err error
	if newConfig == nil {
		return errors.New("Could not compare against config (nil)")
	}

	log.Debugf("Loading namespaces: %s", newConfig.Namespace)
	currentConfig, err := cluster.LoadEnvironment(newConfig.Namespace)

	if err != nil {
		log.Errorf("Error while loading environment: %s", err.Error())
		return err
	}
	if diff.Compare(*newConfig, *currentConfig) {
		err = cluster.ApplyEnvironment(currentConfig, newConfig)
	}

	return err
}

// ApplyEnvironment executes kubectl apply against ingresses, services, deployments
// etc.
func (cluster *Cluster) ApplyEnvironment(currentEnvironment, newEnvironment *bitesize.Environment) error {
	var err error

	for _, service := range newEnvironment.Services {
		if !shouldDeployOnChange(currentEnvironment, newEnvironment, service.Name) {
			continue
		}

		res := bitesize.Imports{}
		// Load configmaps for the service
		for _, vol := range service.Volumes {
			if vol.IsConfigMapVolume() {
				res = append(res, *newEnvironment.Imports.FindByName(vol.Name, bitesize.TypeConfigMap))
			}
		}
		// TODO: load jobs and cronjobs

		err = cluster.ApplyService(&service, &res, newEnvironment.Namespace)
	}
	return err
}

// ApplyService applies a single service to the namespace
func (cluster *Cluster) ApplyService(service *bitesize.Service, imports *bitesize.Imports, namespace string) error {
	var err error
	mapper := &translator.KubeMapper{
		BiteService: service,
		Namespace:   namespace,
		Imports:     imports,
	}

	client := &k8s.Client{
		Interface: cluster.Interface,
		Namespace: namespace,
		CRDClient: cluster.CRDClient,
	}

	if service.Type == "" {
		log.Debugf("Applying pvcs for Service %s ", service.Name)
		pvc, _ := mapper.PersistentVolumeClaims()
		for _, claim := range pvc {
			if err = client.PVC().Apply(&claim); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("Applying configmaps for Service %s ", service.Name)
		cMaps, _ := mapper.ConfigMaps()
		for _, c := range cMaps {
			if err = client.ConfigMap().Apply(&c); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("Applying Deployment for Service %s", service.Name)
		deployment, err := mapper.Deployment()
		if err != nil {
			log.Error(err)
			return err
		}

		if err = client.Deployment().Apply(deployment); err != nil {
			log.Error(err)
		}

		svc, _ := mapper.Service()
		if err = client.Service().Apply(svc); err != nil {
			log.Error(err)
			log.Debugf("Service +%v", svc)
		}

		hpa, _ := mapper.HPA()
		if err = client.HorizontalPodAutoscaler().Apply(hpa); err != nil {
			log.Error(err)
		}

		if service.HasExternalURL() {
			ingress, _ := mapper.Ingress()
			if err = client.Ingress().Apply(ingress); err != nil {
				log.Error(err)
			}
		}

	} else {
		crd, _ := mapper.CustomResourceDefinition()
		if err = client.CustomResourceDefinition(crd.Kind).Apply(crd); err != nil {
			log.Error(err)
		} else {
			log.Infof("Successfully updated CRD resource: %s", crd.Name)
		}
	}
	return err
}

// LoadPods returns Pod object loaded from Kubernetes API
func (cluster *Cluster) LoadPods(namespace string) ([]bitesize.Pod, error) {
	client := &k8s.Client{
		Namespace: namespace,
		Interface: cluster.Interface,
		CRDClient: cluster.CRDClient,
	}

	var deployedPods []bitesize.Pod
	pods, err := client.Pod().List()
	if err != nil {
		log.Errorf("Error loading kubernetes pods: %s", err.Error())
	}

	for _, pod := range pods {
		logs, err := client.Pod().GetLogs(pod.ObjectMeta.Name)
		message := ""
		if err != nil {
			message = fmt.Sprintf("Error retrieving Pod Logs: %s", err.Error())

		}
		podval := bitesize.Pod{
			Name:      pod.ObjectMeta.Name,
			Phase:     pod.Status.Phase,
			StartTime: pod.Status.StartTime.String(),
			Message:   message,
			Logs:      logs,
		}
		deployedPods = append(deployedPods, podval)
	}
	return deployedPods, err
}

// LoadEnvironment returns BitesizeEnvironment object loaded from Kubernetes API
func (cluster *Cluster) LoadEnvironment(namespace string) (*bitesize.Environment, error) {
	serviceMap := make(ServiceMap)

	client := &k8s.Client{
		Namespace: namespace,
		Interface: cluster.Interface,
		CRDClient: cluster.CRDClient,
	}

	ns, err := client.Ns().Get()
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving namespace: %s", err.Error())
	}
	environmentName := ns.ObjectMeta.Labels["environment"]

	services, err := client.Service().List()
	if err != nil {
		log.Errorf("Error loading kubernetes services: %s", err.Error())
	}
	for _, service := range services {
		serviceMap.AddService(service)
	}

	deployments, err := client.Deployment().List()
	if err != nil {
		log.Errorf("Error loading kubernetes deployments: %s", err.Error())
	}
	for _, deployment := range deployments {
		serviceMap.AddDeployment(deployment)
	}

	hpas, err := client.HorizontalPodAutoscaler().List()
	if err != nil {
		log.Errorf("Error loading kubernetes hpas: %s", err.Error())
	}
	for _, hpa := range hpas {
		serviceMap.AddHPA(hpa)
	}

	ingresses, err := client.Ingress().List()
	if err != nil {
		log.Errorf("Error loading kubernetes ingresses: %s", err.Error())
	}

	for _, ingress := range ingresses {
		serviceMap.AddIngress(ingress)
	}

	// we'll need the same for tprs
	claims, _ := client.PVC().List()
	for _, claim := range claims {
		serviceMap.AddVolumeClaim(claim)
	}

	for _, supported := range k8_extensions.SupportedThirdPartyResources {
		crds, _ := client.CustomResourceDefinition(supported).List()
		for _, crd := range crds {
			serviceMap.AddCustomResourceDefinition(crd)
		}
	}

	// Handle imported resources
	resourceMap := new(ResourceMap)

	configmaps, _ := client.ConfigMap().List()
	for _, config := range configmaps {
		resourceMap.AddConfigMap(config)
	}

	bitesizeConfig := bitesize.Environment{
		Name:      environmentName,
		Namespace: namespace,
		Services:  serviceMap.Services(),
		Imports:   resourceMap.Resources(),
	}

	return &bitesizeConfig, nil
}

//Only deploy k8s resources when the environment was actually deployed and changed or if the service has specified a version
func shouldDeployOnChange(currentEnvironment, newEnvironment *bitesize.Environment, serviceName string) bool {
	if !diff.ServiceChanged(serviceName) {
		return false
	}

	currentService := currentEnvironment.Services.FindByName(serviceName)
	updatedService := newEnvironment.Services.FindByName(serviceName)

	if updatedService == nil {
		return true
	}

	if updatedService.IsBlueGreenParentDeployment() {
		log.Debugf("Should deploy blue/green service %s", serviceName)
		return true
	}

	if currentService != nil && currentService.Status.DeployedAt != "" {
		return true
	}

	if updatedService.Version != "" {
		return true
	}
	return false
}
