package bitesize

import (
	"fmt"
	"io/ioutil"

	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// EnvironmentsBitesize is a 1:1 mapping to environments.bitesize file
type EnvironmentsBitesize struct {
	Project      string       `yaml:"project"`
	Environments Environments `yaml:"environments"`
}

// DeploymentSettings represent "deployment" block in environments.bitesize
type DeploymentSettings struct {
	Method string `yaml:"method,omitempty" validate:"regexp=^(bluegreen|rolling-upgrade)*$"`
	Mode   string `yaml:"mode,omitempty" validate:"regexp=^(manual|auto)*$"`
	Active string `yaml:"active,omitempty" validate:"regexp=^(blue|green)*$"`
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

// Metric maps to HPA targets in kubernetes
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
	// Name of the referent.
	Name string `yaml:"name"`
	// Pathe
	Path  string `yaml:"path"`
	Modes string `yaml:"modes" validate:"volume_modes"`
	Size  string `yaml:"size"`
	Type  string `yaml:"type"`
	// If unspecified, each key-value pair in the Data field of the referenced
	// ConfigMap will be projected into the volume as a file whose name is the
	// key and content is the value. If specified, the listed keys will be
	// projected into the specified paths, and unlisted keys will not be
	// present. If a key is specified which is not present in the ConfigMap,
	// the volume setup will error unless it is marked optional. Paths must be
	// relative and may not contain the '..' path or start with '..'.
	// +optional
	Items        []KeyToPath `yaml:"item"`
	provisioning string      `yaml:"provisioning" validate:"volume_provisioning"`
}

// KeyToPath Maps a string key to a path within a volume.
type KeyToPath struct {
	// The key to project.
	Key string `yaml:"key"`
	// The relative path of the file to map the key to.
	// May not be an absolute path.
	// May not contain the path element '..'.
	// May not start with the string '..'.
	Path string `yaml:"path"`
	// Optional: mode bits to use on this file, must be a value between 0
	// and 0777. If not specified, the volume defaultMode will be used.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	// +optional
	Mode *int32 `yaml:"mode,omitempty"`
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

	// TODO: Validate
	//  if err = validator.Validate(e); err != nil {
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
