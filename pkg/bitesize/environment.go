package bitesize

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
	"github.com/pearsontechnology/environment-operator/pkg/util"

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
func LoadEnvironment(pathToBitesizeFile, envName string) (*Environment, error) {
	e, err := LoadFromFile(pathToBitesizeFile)
	if err != nil {
		return nil, err
	}
	util.LogTraceAsYaml("LoadFromFile", e)
	for _, env := range e.Environments {
		if env.Name == envName {
			// Environment name found check for git configs
			rootPath := config.Env.GitLocalPath
			if len(env.Repo.Remote) > 0 {
				gitClient, err = git.EnvGitClient(env.Repo.Remote,
					env.Repo.Branch, env.Namespace, env.Name)
				if err != nil {
					return nil, err
				}
				if err := gitClient.Refresh(); err != nil {
					log.Tracef("Git Client Refresh failed:")
					return nil, fmt.Errorf("  RemotePath=%s"+
						"   LocalPath=%s"+
						"   Branch=%s"+
						"   SSHkey=%s"+
						"   GITUser=%s\n",
						gitClient.RemotePath,
						gitClient.LocalPath,
						gitClient.BranchName,
						gitClient.SSHKey,
						gitClient.GitUser,
					)
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
			util.LogTraceAsYaml("bitesize.LoadEnvironment env", env)
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment %s not found in %s", envName, pathToBitesizeFile)
}

func loadServices(env Environment) Services {
	// load services from an environment
	// Specifies their defaults and handles overrides of user-supplied config
	var blueGreenServices Services
	for i, svc := range env.Services {
		if svc.IsBlueGreenParentDeployment() {
			blueGreenServices = append(blueGreenServices, copyBlueGreenService(svc, BlueService))
			blueGreenServices = append(blueGreenServices, copyBlueGreenService(svc, GreenService))
		}

		// Internally prepend env vars that come from secrets w/o name prefix
		// All env from secrets in .bitesize files must be specified
		// as name/key to remove this.
		for j, envVar := range svc.EnvVars {
			if envVar.Secret != "" && !strings.Contains(envVar.Value, "/") {
				newValue := fmt.Sprintf("%s/%s", envVar.Value, envVar.Value)
				log.Tracef("Service %s: Internally converting Secret %s Value from %s to %s",
					svc.Name, envVar.Secret, envVar.Value, newValue)
				env.Services[i].EnvVars[j].Value = newValue
			}
		}

		// allow config file to specify any type to any letter case.
		// e.g., "EFS" is stored as "efs"
		for j, vol := range svc.Volumes {
			lowerVolType := strings.ToLower(vol.Type)
			if lowerVolType != vol.Type {
				log.Tracef("Service: %s: Internally converting vol.Type from %s to %s",
					svc.Name, vol.Type, lowerVolType)
				env.Services[i].Volumes[j].Type = lowerVolType
			}
		}

		util.LogTraceAsYaml("Unsorted vols", env.Services[i].Volumes)
		vols, err := SortVolumesByVolName(env.Services[i].Volumes)
		util.LogTraceAsYaml("Sorted vols", vols)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sortVolumesByVolName error: %v\n", err)
			os.Exit(1)
		}
		if len(vols) > 0 {
			env.Services[i].Volumes = vols
		}
	}

	services := append(env.Services, blueGreenServices...)
	util.LogTraceAsYaml("All services post-modification", services)
	return services
}

// Sorts volumes by name so they can be put in config file in any order.
// This allows the diff function to be clean when the volums are out of order.
func SortVolumesByVolName(m []Volume) ([]Volume, error) {
	keys := make([]string, 0, len(m))
	for _, v := range m {
		keys = append(keys, v.Name)
	}
	sort.Strings(keys)
	newOrder := make([]Volume, 0, len(keys))
	for _, k := range keys {
		record, err := findVolByName(m, k)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		newOrder = append(newOrder, record)
	}
	return newOrder, nil
}

// Returns a Volume by its name
func findVolByName(m []Volume, name string) (Volume, error) {
	for _, entry := range m {
		if entry.Name == name {
			return entry, nil
		}
	}
	return Volume{}, errors.New("'Name' not found in any volume field")
}
