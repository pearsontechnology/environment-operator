package web

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	v1beta2_apps "k8s.io/api/apps/v1beta2"
	v1beta1_ext "k8s.io/api/extensions/v1beta1"
)

// GetCurrentDeploymentByName retrieves kubernetes deployment object for
// currently active environment from bitesize file in git.
func GetCurrentDeploymentByName(name string) (*v1beta1_ext.Deployment, *v1beta2_apps.StatefulSet, error) {
	gitClient := git.Client()
	gitClient.Refresh()

	environment, err := bitesize.LoadEnvironmentFromConfig(config.Env)
	if err != nil {
		log.Errorf("Could not load env: %s", err.Error())
		return nil, nil, err
	}

	log.Debugf("ENV: %+v", *environment)

	service := environment.Services.FindByName(name)
	if service == nil {
		log.Infof("Services: %v", environment.Services)
		return nil, nil, fmt.Errorf("%s not found", name)
	}

	mapper := translator.KubeMapper{
		BiteService: service,
	}

	if service.DatabaseType == "mongo" {
		statefulset, _ := mapper.MongoStatefulSet()
		if err != nil {
			log.Errorf("Could not process statefulset: %s", err.Error())
			return nil, nil, err
		}
		return nil, statefulset, nil

	} else {
		deployment, err := mapper.Deployment()
		if err != nil {
			log.Errorf("Could not process deployment : %s", err.Error())
			return nil, nil, err
		}
		return deployment, nil, nil
	}
}
