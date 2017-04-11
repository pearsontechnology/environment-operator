package web

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
	"github.com/pearsontechnology/environment-operator/pkg/translator"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// GetCurrentDeploymentByName retrieves kubernetes deployment object for
// currently active environment from bitesize file in git.
func GetCurrentDeploymentByName(name string) (*v1beta1.Deployment, error) {
	gitClient := git.Client()
	gitClient.Refresh()

	cfg := config.Load()
	environment, err := bitesize.LoadEnvironmentFromConfig(cfg)
	if err != nil {
		log.Errorf("Could not load env: %s", err.Error())
		return nil, err
	}

	log.Debugf("ENV: %+v", *environment)

	service := environment.Services.FindByName(name)
	if service == nil {
		log.Infof("Services: %q", environment.Services)
		return nil, fmt.Errorf("%s not found", name)
	}

	mapper := translator.KubeMapper{
		BiteService: service,
	}
	return mapper.Deployment()
}
