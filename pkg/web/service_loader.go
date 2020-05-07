package web

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
)

func loadServiceFromConfig(name string) (*bitesize.Service, error) {
	gitClient := git.Client()
	gitClient.Refresh()

	environment, err := bitesize.LoadEnvironmentFromConfig(config.Env)
	if err != nil {
		return nil, fmt.Errorf("Could not load env: %s", err.Error())
	}

	service := environment.Services.FindByName(name)
	if service == nil {
		log.Warnf("Services: %v", environment.Services)
		return nil, fmt.Errorf("%s not found", name)
	}
	return service, nil
}

func loadServiceFromCluster(name string) (bitesize.Service, error) {
	client, err := cluster.Client()
	if err != nil {
		return bitesize.Service{}, errors.New(fmt.Sprintf("Error cluster client: %s", err.Error()))
	}

	e, err := client.ScrapeResourcesForNamespace(config.Env.Namespace)
	if err != nil {
		return bitesize.Service{}, errors.New(fmt.Sprintf("Error getting environment: %s", err.Error()))
	}

	s := e.Services.FindByName(name)
	if s == nil {
		return bitesize.Service{}, errors.New("Error getting service: name")
	}
	return *s, nil
}

func loadConfigMapsFromConfig() (*bitesize.Gists, error) {
	gitClient := git.Client()
	gitClient.Refresh()

	environment, err := bitesize.LoadEnvironmentFromConfig(config.Env)
	if err != nil {
		return nil, fmt.Errorf("Could not load env: %s", err.Error())
	}

	res := environment.Gists.FindByType(bitesize.TypeConfigMap)
	if res == nil {
		log.Warnf("Imported Resources: %v", res)
		return nil, fmt.Errorf("ConfigMaps not found")
	}
	return &res, nil
}
