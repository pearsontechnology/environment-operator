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

	if diff.Compare(*newConfig, *currentConfig) {

		log.Infof("Changes:\n %s", diff.Changes())
		err = cluster.ApplyEnvironment(currentConfig, newConfig)
	}
	return err
}

// ApplyEnvironment executes kubectl apply against ingresses, services, deployments
// etc.
func (cluster *Cluster) ApplyEnvironment(currentEnvironment, newEnvironment *bitesize.Environment) error {
	var err error

	for _, service := range newEnvironment.Services {

		mapper := &translator.KubeMapper{
			BiteService: &service,
			Namespace:   newEnvironment.Namespace,
		}

		client := &k8s.Client{
			Interface: cluster.Interface,
			Namespace: newEnvironment.Namespace,
			TPRClient: cluster.TPRClient,
		}

		if service.Type == "" {

			if !shouldDeploy(currentEnvironment, newEnvironment, service.Name) {
				continue
			}

			if service.DatabaseType == "mongo" {
				log.Debugf("Applying Stateful set for Mongo DB Service: %s ", service.Name)

				secret, _ := mapper.MongoInternalSecret()

				//Only apply the secret if it doesnt exist. Changing this secret would cause a deployed mongo
				//cluster from being able to communicate between replicas.  Need a way to update this secret
				// and redploy the mongo statefulset. For now, just protect against changing the secret
				// via environment operator
				if !client.Secret().Exists(secret.Name) {
					if err = client.Secret().Apply(secret); err != nil {
						log.Error(err)
					}
				}

				statefulset, _ := mapper.MongoStatefulSet()
				if err = client.StatefulSet().Apply(statefulset); err != nil {
					log.Error(err)
				}

				svc, _ := mapper.HeadlessService()
				if err = client.Service().Apply(svc); err != nil {
					log.Error(err)
				}

			} else { //Only apply a Deployment and PVCs if this is not a DB service. The DB Statefulset creates its own PVCs
				log.Debugf("Applying Deployment for Service %s ", service.Name)
				deployment, err := mapper.Deployment()
				if err != nil {
					return err
				}
				if err = client.Deployment().Apply(deployment); err != nil {
					log.Error(err)
				}

				pvc, _ := mapper.PersistentVolumeClaims()
				for _, claim := range pvc {
					if err = client.PVC().Apply(&claim); err != nil {
						log.Error(err)
					}
				}

				svc, _ := mapper.Service()
				if err = client.Service().Apply(svc); err != nil {
					log.Error(err)
				}
			}

			hpa, _ := mapper.HPA()
			if err = client.HorizontalPodAutoscaler().Apply(&hpa); err != nil {
				log.Error(err)
			}

			if service.HasExternalURL() {
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

// LoadPods returns Pod object loaded from Kubernetes API
func (cluster *Cluster) LoadPods(namespace string) ([]bitesize.Pod, error) {
	client := &k8s.Client{
		Namespace: namespace,
		Interface: cluster.Interface,
		TPRClient: cluster.TPRClient,
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

	statefulsets, err := client.StatefulSet().List()
	if err != nil {
		log.Errorf("Error loading kubernetes statefulsets : %s", err.Error())
	}

	for _, statefulset := range statefulsets {
		serviceMap.AddMongoStatefulSet(statefulset)
	}

	// we'll need the same for tprs
	claims, _ := client.PVC().List()
	for _, claim := range claims {
		serviceMap.AddVolumeClaim(claim)
	}

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

//Only deploy k8s resources when the environment was actually deployed and changed or if the service has specified a version
func shouldDeploy(currentEnvironment, newEnvironment *bitesize.Environment, serviceName string) bool {
	currentService := currentEnvironment.Services.FindByName(serviceName)
	updatedService := newEnvironment.Services.FindByName(serviceName)

	if (currentService != nil && currentService.Status.DeployedAt != "") || (updatedService != nil && updatedService.Version != "") {
		if diff.ServiceChanged(serviceName) {
			return true
		}
	}
	return false
}
