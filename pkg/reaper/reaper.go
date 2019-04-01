package reaper

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"
)

// Reaper goes through orphan objects defined in Namespace and deletes them
type Reaper struct {
	Wrapper   *cluster.Cluster
	Namespace string
}

// Cleanup collects all orphan services or service components (not mentioned in cfg) and
// deletes them from the cluster
func (r *Reaper) Cleanup(cfg *bitesize.Environment) error {

	if cfg == nil || cfg.Services == nil {
		return errors.New("REAPER Error with bitesize file, configuration is nil")
	}

	current, err := r.Wrapper.LoadEnvironment(r.Namespace)
	if err != nil {
		return fmt.Errorf("REAPER Error loading environment: %s", err.Error())
	}

	for _, service := range current.Services {
		configService := cfg.Services.FindByName(service.Name)

		if configService == nil {
			log.Infof("REAPER: Found orphan service %s, deleting.", service.Name)
			r.deleteService(service)
		} else if configService.IsBlueGreenParentDeployment() {
			r.destroyDeployment(service.Name)
		}

		// delete ingresses that were removed from the service config
		r.CleanupIngress(cfg.Services.FindByName(service.Name), &service)
	}
	return nil
}

// deleteService removes deployments, ingresses, services and crds related to
// the bitesize service from the cluster
func (r *Reaper) deleteService(svc bitesize.Service) error {

	r.destroyIngress(svc.Name)
	r.destroyDeployment(svc.Name)
	r.destroyService(svc.Name)
	for _, volume := range svc.Volumes {
		r.destroyPersistentVolume(volume.Name)
	}
	r.destroyCustomResourceDefinition(svc.Name)
	return nil
}

// XXX: I hate this repetition

func (r *Reaper) destroyIngress(name string) error {
	client := k8s.Ingress{
		Interface: r.Wrapper.Interface,
		Namespace: r.Namespace,
	}

	return client.Destroy(name)
}

func (r *Reaper) destroyDeployment(name string) error {
	client := k8s.Deployment{
		Interface: r.Wrapper.Interface,
		Namespace: r.Namespace,
	}
	if client.Exist(name) {
		return client.Destroy(name)
	}
	return nil
}

func (r *Reaper) destroyService(name string) error {
	client := k8s.Service{
		Interface: r.Wrapper.Interface,
		Namespace: r.Namespace,
	}
	return client.Destroy(name)
}

func (r *Reaper) destroyPersistentVolume(name string) error {
	client := k8s.PersistentVolumeClaim{
		Interface: r.Wrapper.Interface,
		Namespace: r.Namespace,
	}
	return client.Destroy(name)
}

func (r *Reaper) destroyCustomResourceDefinition(name string) error {
	return nil
}

// CleanupIngress deletes an ingress if the corresponding service external_url is removed from the config
func (r *Reaper) CleanupIngress(configSvc, clusterSvc *bitesize.Service) {
	if configSvc != nil && !configSvc.HasExternalURL() && clusterSvc.HasExternalURL() {
		log.Infof("REAPER: deleting ingress %s because it was removed from the service config", clusterSvc.Name)
		r.destroyIngress(clusterSvc.Name)
	}
}
