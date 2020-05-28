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
	Method     string              `yaml:"method,omitempty" validate:"regexp=^(bluegreen|rolling-upgrade)*$"`
	Mode       string              `yaml:"mode,omitempty" validate:"regexp=^(manual|auto)*$"`
	BlueGreen  *BlueGreenSettings  `yaml:"-"`
	CustomURLs map[string][]string `yaml:"custom_urls,omitempty"`
	// XXX    map[string]interface{} `yaml:",inline"`
}

// BlueGreenSettings is a collection of internal bluegreen settings
// used as a various helpers in deployment
type BlueGreenSettings struct {
	Active           *BlueGreenServiceSet // used in "parent" service to determine which environment is active
	DeploymentColour *BlueGreenServiceSet // used in "child" blue/green service to indicate it's colour
	ActiveFlag       bool                 // used in "child" blue/green service to indicate whethen this environment is currently active
}

// HorizontalPodAutoscaler maps to HPA in kubernetes
type HorizontalPodAutoscaler struct {
	MinReplicas int32  `yaml:"min_replicas"`
	MaxReplicas int32  `yaml:"max_replicas"`
	Metric      Metric `yaml:"metric"`
}

// Container maps a single application container that you want to run within a pod
type Container struct {
	Application string   `yaml:"application,omitempty"`
	Name        string   `yaml:"name" validate:"nonzero"`
	Version     string   `yaml:"version,omitempty"`
	EnvVars     []EnvVar `yaml:"env,omitempty"`
	Command     []string `yaml:"command"`
	Volumes     []Volume `yaml:"volumes,omitempty"`
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

// Probe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
type Probe struct {
	Handler             `yaml:"handler"`
	InitialDelaySeconds int32 `yaml:"initial_delay_seconds,omitempty"`
	TimeoutSeconds      int32 `yaml:"timeout_seconds,omitempty"`
	PeriodSeconds       int32 `yaml:"period_seconds,omitempty"`
	SuccessThreshold    int32 `yaml:"success_threshold,omitempty"`
	FailureThreshold    int32 `yaml:"failure_threshold,omitempty"`
}

type Handler struct {
	Exec      *ExecAction      `yaml:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `yaml:"http_get,omitempty"`
	TCPSocket *TCPSocketAction `yaml:"tcp_socket,omitempty"`
}

type ExecAction struct {
	Command []string `yaml:"command,omitempty"`
}

type HTTPGetAction struct {
	Path        string       `yaml:"path,omitempty"`
	Port        int32        `yaml:"port"`
	Host        string       `yaml:"host,omitempty"`
	Scheme      v1.URIScheme `yaml:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `yaml:"http_headers,omitempty"`
}

type HTTPHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type TCPSocketAction struct {
	Port int32  `yaml:"port"`
	Host string `yaml:"host,omitempty"`
}

// EnvVar represents environment variables in pod
type EnvVar struct {
	Name     string `yaml:"name,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Secret   string `yaml:"secret,omitempty"`
	PodField string `yaml:"pod_field,omitempty"`
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
	// Path
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
	Items []KeyToPath `yaml:"items"`
	// volume provisioning types accepted 'dynamic' and 'manual'
	provisioning string `yaml:"provisioning" validate:"volume_provisioning"`
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
	var a struct {
		Active *BlueGreenServiceSet `yaml:"active,omitempty"`
	}

	ee := &DeploymentSettings{Method: "rolling-upgrade"}
	type plain DeploymentSettings
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("deployment.%s", err.Error())
	}

	if err := unmarshal(&a); err != nil {
		return fmt.Errorf("deployment.%s", err.Error())
	}

	*e = *ee

	if a.Active != nil {
		e.BlueGreen = &BlueGreenSettings{Active: a.Active}
	}

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
