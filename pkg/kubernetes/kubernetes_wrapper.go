package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
)

// Wrapper wraps low level kubernetes api requests to an object easier
// to interact with
type Cluster struct {
	kubernetes.Interface
}

// NewWrapper returns default in-cluster kubernetes client
func NewClusterClient() (*Cluster, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Cluster{Interface: clientset}, nil
}

// ApplyEnvironment executes kubectl apply against ingresses, services, deployments
// etc.
func (cluster *Cluster) ApplyEnvironment(e *bitesize.Environment) {
	var err error

	for _, service := range e.Services {

		mapper := &translator.KubeMapper{
			BiteService: &service,
			Namespace:   e.Namespace,
		}

		client := &k8s.Client{
			Interface: cluster.Interface,
			Namespace: e.Namespace,
		}

		if service.Type == "" {
			svc, _ := mapper.Service()
			if err = client.Service().Apply(svc); err != nil {
				log.Error(err)
			}

			deployment, _ := mapper.Deployment()
			if err = client.Deployment().Apply(deployment); err != nil {
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
}
