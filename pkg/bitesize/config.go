package bitesize

import (
	"fmt"
	"io/ioutil"

	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

// EnvironmentsBitesize is a 1:1 mapping to environments.bitesize file
type EnvironmentsBitesize struct {
	Project      string       `yaml:"project"`
	Environments Environments `yaml:"environments"`
	// XXX          map[string]interface{} `yaml:",inline"`
}

// DeploymentSettings represent "deployment" block in environments.bitesize
type DeploymentSettings struct {
	Method string `yaml:"method,omitempty" validate:"regexp=^(bluegreen|rolling-upgrade)*$"`
	Mode   string `yaml:"mode,omitempty" validate:"regexp=^(manual|auto)*$"`
	Active string `yaml:"active,omitempty" validate:"regexp=^(blue|green)*$"`
	// XXX    map[string]interface{} `yaml:",inline"`
}

// HorizontalPodAutoscaler maps to HPA in kubernetes
type HorizontalPodAutoscaler struct {
	MinReplicas int32  `yaml:"min_replicas"`
	MaxReplicas int32  `yaml:"max_replicas"`
	Metric      Metric `yaml:"metric"`
}

// ContainerRequests maps to requests in kubernetes
type ContainerRequests struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// ContainerLimits maps to limits in kubernetes
type ContainerLimits struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// Metrics maps to HPA targets in kubernetes
type Metric struct {
	Name                     string `yaml:"name"`
	TargetAverageValue       string `yaml:"target_average_value,omitempty"`
	TargetAverageUtilization int32  `yaml:"target_average_utilization,omitempty"`
}

// Test is obsolete and not used by environment-operator,
// but it's here for configuration compatability
type Test struct {
	Name       string              `yaml:"name"`
	Repository string              `yaml:"repository"`
	Branch     string              `yaml:"branch"`
	Commands   []map[string]string `yaml:"commands"`
	// XXX        map[string]interface{} `yaml:",inline"`
}

// HealthCheck maps to LivenessProbe in Kubernetes
type HealthCheck struct {
	Command      []string `yaml:"command"`
	InitialDelay int      `yaml:"initial_delay,omitempty"`
	Timeout      int      `yaml:"timeout,omitempty"`
	// XXX          map[string]interface{} `yaml:",inline"`
}

// EnvVar represents environment variables in pod
type EnvVar struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"value"`
	Secret   string `yaml:"secret"`
	PodField string `yaml:"pod_field"`
}

// Pod represents Pod in Kubernetes
type Pod struct {
	Name      string      `yaml:"name"`
	Phase     v1.PodPhase `yaml:"phase"`
	StartTime string      `yaml:"start_time"`
	Message   string      `yaml:"message"`
	Logs      string      `yaml:"logs"`
}

// Annotation represents annotation variables in pod
type Annotation struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// Volume represents volume & it's mount
type Volume struct {
	Name         string `yaml:"name"`
	Path         string `yaml:"path"`
	Modes        string `yaml:"modes" validate:"volume_modes"`
	Size         string `yaml:"size"`
	Type         string `yaml:"type"`
	provisioning string `yaml:"provisioning" validate:"volume_provisioning"`
}

func init() {
	addCustomValidators()
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for HealthCHeck.
func (e *HealthCheck) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	ee := &HealthCheck{}
	type plain HealthCheck
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("health_check.%s", err.Error())
	}

	*e = *ee

	// if err = validator.Validate(e); err != nil {
	// 	return fmt.Errorf("health_check.%s", err.Error())
	// }
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for DeploymentSettings.
func (e *DeploymentSettings) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	ee := &DeploymentSettings{}
	type plain DeploymentSettings
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("deployment.%s", err.Error())
	}

	*e = *ee
	if err = validator.Validate(e); err != nil {
		return fmt.Errorf("deployment.%s", err.Error())
	}
	return nil
}

// LoadFromString returns BitesizeEnvironment object from yaml string
func LoadFromString(cfg string) (*EnvironmentsBitesize, error) {
	t := &EnvironmentsBitesize{}
	err := yaml.Unmarshal([]byte(cfg), t)
	return t, err
}

// LoadFromFile returns BitesizeEnvironment object loaded from file, passed
// as a path argument.
func LoadFromFile(path string) (*EnvironmentsBitesize, error) {
	var err error
	var contents []byte

	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromString(string(contents))
}

// func checkOverflow(m map[string]interface{}, ctx string) error {
// 	if len(m) > 0 {
// 		var keys []string
// 		for k := range m {
// 			keys = append(keys, k)
// 		}
// 		return fmt.Errorf("%s: unknown fields (%s)", ctx, strings.Join(keys, ", "))
// 	}
// 	return nil
// }
