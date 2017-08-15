package bitesize

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/pearsontechnology/environment-operator/pkg/config"

	validator "gopkg.in/validator.v2"
)

// Environment represents full managed environment,
// including services, HTTP endpoints and deployments. It can
// be either built from environments.bitesize configuration file
// or Kubernetes cluster
type Environment struct {
	Name       string              `yaml:"name" validate:"nonzero"`
	Namespace  string              `yaml:"namespace,omitempty" validate:"regexp=^[a-zA-Z\\-]*$"` // This field should be optional now
	Deployment *DeploymentSettings `yaml:"deployment,omitempty"`
	Services   Services            `yaml:"services"`
	Tests      []Test              `yaml:"tests,omitempty"`
	// XXX        map[string]interface{} `yaml:",inline"`
}

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

// LoadEnvironment returns bitesize.Environment object
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

	validator.SetValidationFunc("volume_modes", validVolumeModes)
	validator.SetValidationFunc("hpa", validHPA)

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
			return &env, nil
		}
	}
	return nil, fmt.Errorf("Environment %s not found in %s", envName, path)
}
