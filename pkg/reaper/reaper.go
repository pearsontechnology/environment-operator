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
		return errors.New("REAPER: error with bitesize file, configuration is nil")
	}

	current, err := r.Wrapper.LoadEnvironment(r.Namespace)
	if err != nil {
		return fmt.Errorf("REAPER: error loading environment: %s", err.Error())
	}

	for _, service := range current.Services {
		configService := cfg.Services.FindByName(service.Name)

		if configService == nil {
			log.Infof("REAPER: found orphan service %s, deleting.", service.Name)
			err := r.deleteService(service)
			log.Error(err)
		} else if configService.IsBlueGreenParentDeployment() {
			err := r.destroyDeployment(service.Name)
			log.Error(err)
		}

		// delete ingresses that were removed from the service config
		r.CleanupIngress(cfg.Services.FindByName(service.Name), &service)
	}

	// cleanup all resources that were removed from the service config
	r.CleanupGists(cfg.Gists, current.Gists)

	return nil
}

// deleteService removes deployments, ingresses, services and crds related to
// the BiteSize service from the cluster
func (r *Reaper) deleteService(svc bitesize.Service) error {

	if err := r.destroyIngress(svc.Name); err != nil {
		log.Errorf("REAPER: failed to destroy ingress: %s", err.Error())
	}

	if err := r.destroyDeployment(svc.Name); err != nil {
		log.Errorf("REAPER: failed to destroy deployment: %s",err.Error())
	}

	if err := r.destroyService(svc.Name); err != nil {
		log.Errorf("REAPER: failed to destroy service failed: %s",err.Error())
	}

	for _, volume := range svc.Volumes {
		if err := r.destroyPersistentVolume(volume.Name); err != nil {
			log.Errorf("REAPER: failed to destroy persistent volume: %s",err.Error())
		}
	}

	if err := r.destroyCustomResourceDefinition(svc.Name); err != nil {
		log.Errorf("REAPER: failed to destroy custom resources: %s", err.Error())
	}
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

func (r *Reaper) destroyResource(name string, rstype string) error {
	switch rstype {
	case bitesize.TypeConfigMap:
		{
			client := k8s.ConfigMap{
				Interface: r.Wrapper.Interface,
				Namespace: r.Namespace,
			}
			return client.Destroy(name)
		}
	case bitesize.TypeJob:
		{
			client := k8s.Job{
				Interface: r.Wrapper.Interface,
				Namespace: r.Namespace,
			}
			return client.Destroy(name)
		}
	case bitesize.TypeCronJob:
		{
			client := k8s.CronJob{
				Interface: r.Wrapper.Interface,
				Namespace: r.Namespace,
			}
			return client.Destroy(name)
		}
	}

	return nil
}

// CleanupIngress deletes an ingress if the corresponding service external_url is removed from the config
func (r *Reaper) CleanupIngress(configSvc, clusterSvc *bitesize.Service) {
	if configSvc != nil && !configSvc.HasExternalURL() && clusterSvc.HasExternalURL() {
		log.Infof("REAPER: deleting ingress %s because it was removed from the service config", clusterSvc.Name)
		err := r.destroyIngress(clusterSvc.Name)
		if err != nil {
			log.Error(err)
		}
	}
}

// CleanupGists deletes all gist types imported, if the corresponding gist is removed from the config
func (r *Reaper) CleanupGists(configRes bitesize.Gists, clusterRes bitesize.Gists) {
	if configRes != nil {
		for _, res := range clusterRes {
			found := false
			for _, cfgRes := range configRes {
				if res.Name == cfgRes.Name {
					found = true
				}
			}
			if !found {
				log.Infof("REAPER: Found orphan resource %s, type %s deleting.", res.Name, res.Type)
				err := r.destroyResource(res.Name, res.Type)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
}
