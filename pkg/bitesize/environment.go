package bitesize

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"

	"gopkg.in/validator.v2"
)

// Environment represents full managed environment,
// including services, HTTP endpoints and deployments. It can
// be either built from environments.bitesize configuration file
// or Kubernetes cluster
type Environment struct {
	Name       string              `yaml:"name" validate:"nonzero"`
	Namespace  string              `yaml:"namespace,omitempty" validate:"regexp=^[a-zA-Z0-9\\-]*$"` // This field should be optional now
	Deployment *DeploymentSettings `yaml:"deployment,omitempty"`
	Services   Services            `yaml:"services"`
	Tests      []Test              `yaml:"tests,omitempty"`
	Gists      Gists               `yaml:"gists,omitempty"`
	Repo       GistsRepository     `yaml:"gists_repository,omitempty"`
}

var gitClient *git.Git

// Environments is a custom type to implement sort.Interface
type Environments []Environment

func (slice Environments) Len() int {
	return len(slice)
}

func (slice Environments) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice Environments) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// LoadEnvironmentFromConfig returns bitesize.Environment object
// constructed from environment variables
func LoadEnvironmentFromConfig(c config.Config) (*Environment, error) {
	fp := filepath.Join(c.GitLocalPath, c.EnvFile)
	return LoadEnvironment(fp, c.EnvName)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for BitesizeEnvironment.
func (e *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	ee := &Environment{}
	type plain Environment
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("environment.%s", err.Error())
	}

	*e = *ee

	if err = validator.Validate(e); err != nil {
		return fmt.Errorf("environment.%s", err.Error())
	}
	sort.Sort(e.Services)
	return nil
}

// LoadEnvironment loads named environment from a filename with a given path
func LoadEnvironment(path, envName string) (*Environment, error) {
	e, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	for _, env := range e.Environments {
		if env.Name == envName {
			// Environment name found check for git configs
			rootPath := config.Env.GitLocalPath
			if len(env.Repo.Remote) > 0 {
				gitClient, err = git.EnvGitClient(env.Repo.Remote, env.Repo.Branch, env.Namespace, env.Name)
				if err != nil {
					return nil, err
				}
				if err := gitClient.Refresh(); err != nil {
					return nil, fmt.Errorf("git client information: \n RemotePath=%s \n LocalPath=%s \n Branch=%s \n SSHkey= \n %s", gitClient.RemotePath, gitClient.LocalPath, gitClient.BranchName, gitClient.SSHKey)
				}
				rootPath = gitClient.LocalPath
			}
			// load imported resources
			for k, im := range env.Gists {
				err := LoadResource(&im, env.Namespace, rootPath)
				if err != nil {
					return nil, fmt.Errorf("unable to load Gist type %s, in paths %s,%v %s", im.Type, im.Path, im.Files, err.Error())
				}
				env.Gists[k] = im
			}
			env.Services = loadServices(env)
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment %s not found in %s", envName, path)
}

func loadServices(env Environment) Services {
	var blueGreenServices Services
	for _, svc := range env.Services {
		if svc.IsBlueGreenParentDeployment() {
			blueGreenServices = append(blueGreenServices, copyBlueGreenService(svc, BlueService))
			blueGreenServices = append(blueGreenServices, copyBlueGreenService(svc, GreenService))
		}
	}

	return append(env.Services, blueGreenServices...)
}
