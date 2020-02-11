package cluster

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	"github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	crdcli, err := k8s.CRDClient(nil)
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
		return errors.New("could not compare against config (nil)")
	}

	log.Debugf("loading namespaces: %s", newConfig.Namespace)
	currentConfig, err := cluster.LoadEnvironment(newConfig.Namespace)

	if err != nil {
		log.Errorf("error while loading environment: %s", err.Error())
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

		gists := bitesize.Gists{}
		// Load configmaps for the service
		for _, vol := range service.Volumes {
			if vol.IsConfigMapVolume() {
				res := newEnvironment.Gists.FindByName(vol.Name, bitesize.TypeConfigMap)
				if res == nil {
					log.Warnf("could not find import source for the configmap volume %s", vol.Name)
					continue
				}
				gists = append(gists, *res)
			}
		}

		// Load configmaps for the init containers
		if service.InitContainers != nil {
			for _, containers := range *service.InitContainers {
				for _, vol := range containers.Volumes {
					if vol.IsConfigMapVolume() {
						res := newEnvironment.Gists.FindByName(vol.Name, bitesize.TypeConfigMap)
						if res == nil {
							log.Warnf("could not find import source for the configmap volume %s", vol.Name)
							continue
						}
						gists = append(gists, *res)
					}
				}
			}
		}
		// TODO: load jobs and cronjobs

		err = cluster.ApplyService(&service, &gists, newEnvironment.Namespace)
	}
	return err
}

// ApplyService applies a single service to the namespace
func (cluster *Cluster) ApplyService(service *bitesize.Service, gists *bitesize.Gists, namespace string) error {
	var err error
	mapper := &translator.KubeMapper{
		BiteService: service,
		Namespace:   namespace,
		Gists:       gists,
	}

	client := &k8s.Client{
		Interface: cluster.Interface,
		Namespace: namespace,
		CRDClient: cluster.CRDClient,
	}

	if service.Type == "" {
		log.Debugf("applying pvcs for service %s ", service.Name)
		pvc, _ := mapper.PersistentVolumeClaims()
		for _, claim := range pvc {
			log.Debugf("pvc: %s", claim.Name)
			if err = client.PVC().Apply(&claim); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("applying configmaps for service %s ", service.Name)
		cMaps, _ := mapper.ConfigMaps()
		for _, c := range cMaps {
			log.Debugf("configmap: %s", c.Name)
			if err = client.ConfigMap().Apply(&c); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("applying deployment for service %s", service.Name)
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
			log.Debugf("service +%v", svc)
		}

		hpa, _ := mapper.HPA()
		if err = client.HorizontalPodAutoscaler().Apply(hpa); err != nil {
			log.Error(err)
		}

		if service.HasExternalURL() {

			log.Debugf("applying ingress for service %s", service.Name)
			ingress, _ := mapper.Ingress()
			if err = client.Ingress().Apply(ingress); err != nil {
				log.Error(err)
			}

			if k8s.ExternalSecretsEnabled {
				log.Debugf("applying external secret for ingress %s", service.Name)
				if err := createExternalSecret(mapper, *client, ""); err != nil {
					log.Error("Failed to create ExternalSecret")
				}
			}

			if service.IsServiceMeshEnabled() {

				if k8s.ExternalSecretsEnabled {
					if err := createExternalSecret(mapper, *client, "istio-system"); err != nil {
						log.Error("Failed to create ExternalSecret")
					}
				}

				client.CRDClient, err = k8s.CRDClient(&schema.GroupVersion{
					Group:   "networking.istio.io",
					Version: "v1alpha3",
				})

				if err != nil {
					log.Fatalf("Error creating kubernetes client for ServiceMesh use: %s", err.Error())
				}

				gateway, _ := mapper.ServiceMeshGateway()
				if err = client.CustomResourceDefinition("Gateway").Apply(gateway); err != nil {
					log.Error(err)
				} else {
					log.Infof("Successfully updated Gateway CRD resource: %s", gateway.Name)
				}

				virtualService, _ := mapper.ServiceMeshVirtualService()
				if err = client.CustomResourceDefinition("VirtualService").Apply(virtualService); err != nil {
					log.Error(err)
				} else {
					log.Infof("Successfully updated VirtualService CRD resource: %s", gateway.Name)
				}
			}
		}

	} else {
		crd, _ := mapper.CustomResourceDefinition()

		client.CRDClient, err = k8s.CRDClient(&schema.GroupVersion{
			Group:   strings.Split(crd.TypeMeta.APIVersion, "/")[0],
			Version: strings.Split(crd.TypeMeta.APIVersion, "/")[1],
		})

		if err != nil {
			log.Fatalf("Error creating kubernetes client: %s", err.Error())
		}

		if err = client.CustomResourceDefinition(crd.Kind).Apply(crd); err != nil {
			log.Error(err)
		} else {
			log.Infof("successfully updated CRD resource: %s", crd.Name)
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
		return nil, fmt.Errorf("error while retrieving namespace: %s", err.Error())
	}
	environmentName := ns.ObjectMeta.Labels["environment"]

	services, err := client.Service().List()
	if err != nil {
		log.Errorf("error loading kubernetes services: %s", err.Error())
	}
	for _, service := range services {
		serviceMap.AddService(service)
	}

	deployments, err := client.Deployment().List()
	if err != nil {
		log.Errorf("error loading kubernetes deployments: %s", err.Error())
	}
	for _, deployment := range deployments {
		serviceMap.AddDeployment(deployment)
	}

	hpas, err := client.HorizontalPodAutoscaler().List()
	if err != nil {
		log.Errorf("error loading kubernetes hpas: %s", err.Error())
	}
	for _, hpa := range hpas {
		serviceMap.AddHPA(hpa)
	}

	ingresses, err := client.Ingress().List()
	if err != nil {
		log.Errorf("error loading kubernetes ingresses: %s", err.Error())
	}

	for _, ingress := range ingresses {
		serviceMap.AddIngress(ingress)
	}

	// we'll need the same for tprs
	claims, _ := client.PVC().List()
	for _, claim := range claims {
		serviceMap.AddVolumeClaim(claim)
	}

	for _, apis := range k8_extensions.SupportedCustomResourceAPIVersions {
		// This will ensure that CRDClient creation won't happen during the unit tests.
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) != 0 && len(port) != 0 {
			client.CRDClient, err = k8s.CRDClient(&schema.GroupVersion{
				Group:   strings.Split(apis, "/")[0],
				Version: strings.Split(apis, "/")[1],
			})

			if err != nil {
				log.Fatalf("Error creating kubernetes client: %s", err.Error())
			}
		}

		for _, supported := range k8_extensions.SupportedCustomResources {
			crds, _ := client.CustomResourceDefinition(supported).List()
			for _, crd := range crds {
				serviceMap.AddCustomResourceDefinition(crd)
			}
		}
	}

	// Handle imported resources
	gistMap := GistMap{}

	configmaps, _ := client.ConfigMap().List()
	for _, config := range configmaps {
		gistMap.AddConfigMap(config)
	}

	bitesizeConfig := bitesize.Environment{
		Name:      environmentName,
		Namespace: namespace,
		Services:  serviceMap.Services(),
		Gists:     gistMap.Gists(),
	}

	return &bitesizeConfig, nil
}

// Only deploy k8s resources when the environment was actually deployed and changed or if the service has specified a version
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
		log.Debugf("should deploy blue/green service %s", serviceName)
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

func createExternalSecret(mapper *translator.KubeMapper, client k8s.Client, ns string) error {

	es, err := mapper.ExternalSecretTLS()
	if err != nil {
		log.Errorf("Error creating external secret CRD for ingress: %s", err.Error())
		return err
	}

	if ns != "" {
		es.Namespace = ns
		client.Namespace = ns
	}

	client.CRDClient, err = k8s.CRDClient(&schema.GroupVersion{
		Group:   "kubernetes-client.io",
		Version: "v1",
	})

	if err != nil {
		log.Errorf("Error creating kubernetes client for External Secrets use: %s", err.Error())
		return err
	}

	if err = client.ExternalSecret().Apply(es); err != nil {
		log.Fatalf("Error creating external secret CRD for ingress: %s", err.Error())
		return err
	} else {
		log.Infof("Successfully updated ExternalSecret CRD resource: %s", es.Name)
	}

	return nil
}
