package kubernetes

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
)

// ApplyIfChanged compares bitesize Environment passed as an argument to
// the current client environment. If there are any changes, c is applied
// to the current config
func (cluster *Cluster) ApplyIfChanged(newConfig *bitesize.Environment) error {
	log.Infof("Loading namespace: %s", newConfig.Namespace)
	currentConfig, _ := cluster.LoadEnvironment(newConfig.Namespace)

	changes := diff.Compare(*newConfig, *currentConfig)
	if changes != "" {
		log.Infof("Changes: %s", changes)
		cluster.ApplyEnvironment(newConfig)
	}
	return nil
}

// LoadEnvironment returns BitesizeEnvironment object loaded from Kubernetes API
func (cluster *Cluster) LoadEnvironment(namespace string) (*bitesize.Environment, error) {
	serviceMap := make(ServiceMap)

	client := &k8s.Client{
		Namespace: namespace,
		Interface: cluster.Interface,
	}

	ns, err := client.Ns().Get()
	if err != nil {
		return nil, fmt.Errorf("Namespace %s not found", namespace)
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

	bitesizeConfig := bitesize.Environment{
		Name:      environmentName,
		Namespace: namespace,
		Services:  serviceMap.Services(),
	}

	return &bitesizeConfig, nil
}
