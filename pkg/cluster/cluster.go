package cluster

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/Sirupsen/logrus"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
)

// NewClusterClient returns default in-cluster kubernetes client
func Client() (*Cluster, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	tprclient, err := k8s.TPRClient()
	if err != nil {
		return nil, err
	}

	return &Cluster{Interface: clientset, TPRClient: tprclient}, nil
}

// ApplyIfChanged compares bitesize Environment passed as an argument to
// the current client environment. If there are any changes, c is applied
// to the current config
func (cluster *Cluster) ApplyIfChanged(newConfig *bitesize.Environment) error {
	var err error
	if newConfig == nil {
		return errors.New("Could not compare against config (nil)")
	}
	log.Debugf("Loading namespace: %s", newConfig.Namespace)
	currentConfig, _ := cluster.LoadEnvironment(newConfig.Namespace)

	changes := diff.Compare(*newConfig, *currentConfig)
	if changes != "" {
		log.Infof("Changes: %s", changes)
		err = cluster.ApplyEnvironment(newConfig)
	}
	return err
}

// ApplyEnvironment executes kubectl apply against ingresses, services, deployments
// etc.
func (cluster *Cluster) ApplyEnvironment(e *bitesize.Environment) error {
	var err error

	for _, service := range e.Services {

		mapper := &translator.KubeMapper{
			BiteService: &service,
			Namespace:   e.Namespace,
		}

		client := &k8s.Client{
			Interface: cluster.Interface,
			Namespace: e.Namespace,
			TPRClient: cluster.TPRClient,
		}

		if service.Type == "" {
			svc, _ := mapper.Service()
			if err = client.Service().Apply(svc); err != nil {
				log.Error(err)
			}

			deployment, err := mapper.Deployment()
			if err != nil {
				return err
			}

			if err = client.Deployment().Apply(deployment); err != nil {
				log.Error(err)
			}

			hpa, _ := mapper.HPA()
			if err = client.HorizontalPodAutoscaler().Apply(&hpa); err != nil {
				log.Error(err)
			}

			pvc, _ := mapper.PersistentVolumeClaims()
			for _, claim := range pvc {
				if err = client.PVC().Apply(&claim); err != nil {
					log.Error(err)
				}
			}

			if service.ExternalURL != "" {
				ingress, _ := mapper.Ingress()
				if err = client.Ingress().Apply(ingress); err != nil {
					log.Error(err)
				}
			}

		} else {
			tpr, _ := mapper.ThirdPartyResource()
			if err = client.ThirdPartyResource(tpr.Kind).Apply(tpr); err != nil {
				log.Error(err)
			}
		}
	}
	return err
}

// LoadEnvironment returns BitesizeEnvironment object loaded from Kubernetes API
func (cluster *Cluster) LoadEnvironment(namespace string) (*bitesize.Environment, error) {
	serviceMap := make(ServiceMap)

	client := &k8s.Client{
		Namespace: namespace,
		Interface: cluster.Interface,
		TPRClient: cluster.TPRClient,
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

	pods, err := client.Pod().List()
	if err != nil {
		log.Errorf("Error loading kubernetes pods: %s", err.Error())
	}

	for _, pod := range pods {
		logs, err := client.Pod().GetLogs(pod.ObjectMeta.Name)
		message := ""
		if err != nil {
			message = fmt.Sprintf("Error retrieving Pod Logs: %s", err.Error())
			serviceMap.AddPod(pod, logs, message)
		} else {
			serviceMap.AddPod(pod, logs, message)
		}
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

	// We need to exclude this part from unit tests as we cannot
	// fake client.ThirdPartyResource()
	for _, supported := range k8_extensions.SupportedThirdPartyResources {
		tprs, _ := client.ThirdPartyResource(supported).List()
		for _, tpr := range tprs {
			serviceMap.AddThirdPartyResource(tpr)
		}
	}

	bitesizeConfig := bitesize.Environment{
		Name:      environmentName,
		Namespace: namespace,
		Services:  serviceMap.Services(),
	}

	return &bitesizeConfig, nil
}
