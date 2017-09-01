package bitesize

import (
	"fmt"
	"strconv"
	"strings"

	validator "gopkg.in/validator.v2"
)

// Service represents a single service and it's configuration,
// running in environment
type Service struct {
	Name         string                  `yaml:"name" validate:"nonzero"`
	ExternalURL  string                  `yaml:"external_url,omitempty" validate:"regexp=^([a-zA-Z\\.\\-]+$)*"`
	Ports        []int                   `yaml:"-"` // Ports have custom unmarshaler
	Ssl          string                  `yaml:"ssl,omitempty" validate:"regexp=^(true|false)*$"`
	Version      string                  `yaml:"version,omitempty"`
	Application  string                  `yaml:"application,omitempty"`
	Replicas     int                     `yaml:"replicas,omitempty"`
	Deployment   *DeploymentSettings     `yaml:"deployment,omitempty"`
	HPA          HorizontalPodAutoscaler `yaml:"hpa" validate:"hpa"`
	Requests     ContainerRequests       `yaml:"requests" validate:"requests"`
	HealthCheck  *HealthCheck            `yaml:"health_check,omitempty"`
	EnvVars      []EnvVar                `yaml:"env,omitempty"`
	DeployedPods []Pod                   `yaml:"-"` //Ignore field when parsing bitesize yaml
	Annotations  map[string]string       `yaml:"-"` // Annotations have custom unmarshaler
	Volumes      []Volume                `yaml:"volumes,omitempty"`
	Options      map[string]string       `yaml:"options,omitempty"`
	HTTPSOnly    string                  `yaml:"httpsOnly,omitempty" validate:"regexp=^(true|false)*$"`
	HTTPSBackend string                  `yaml:"httpsBackend,omitempty" validate:"regexp=^(true|false)*$"`
	Type         string                  `yaml:"type,omitempty"`
	Status       ServiceStatus           `yaml:"status,omitempty"`
	// XXX          map[string]interface{} `yaml:",inline"`
}

type ServiceStatus struct {
	DeployedAt        string
	AvailableReplicas int
	DesiredReplicas   int
	CurrentReplicas   int
}

// Services implement sort.Interface
type Services []Service

// ServiceWithDefaults returns new *Service object with default values set
func ServiceWithDefaults() *Service {
	return &Service{
		Ports:    []int{80},
		Replicas: 1,
	}
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for BitesizeService.
func (e *Service) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	ee := ServiceWithDefaults()

	ports, err := unmarshalPorts(unmarshal)
	if err != nil {
		return fmt.Errorf("service.ports.%s", err.Error())
	}

	annotations, err := unmarshalAnnotations(unmarshal)
	if err != nil {
		return fmt.Errorf("service.annotations.%s", err.Error())
	}

	type plain Service
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("service.%s", err.Error())
	}

	*e = *ee
	e.Ports = ports
	e.Annotations = annotations

	if e.Type != "" {
		e.Ports = nil
	}

	// annotation := Annotation{Name: "Name", Value: e.Name}
	// e.Annotations = append(e.Annotations, annotation)

	if e.HPA.MinReplicas != 0 {
		e.Replicas = int(e.HPA.MinReplicas)
	}

	if err = validator.Validate(e); err != nil {
		return fmt.Errorf("service.%s", err.Error())
	}

	return nil
}

func (slice Services) Len() int {
	return len(slice)
}

func (slice Services) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice Services) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// FindByName returns service with a name match
func (slice Services) FindByName(name string) *Service {
	for _, s := range slice {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func unmarshalAnnotations(unmarshal func(interface{}) error) (map[string]string, error) {
	// annotations representation in environments.bitesize
	var bz struct {
		Annotations []struct {
			Name  string
			Value string
		} `yaml:"annotations,omitempty"`
	}
	annotations := map[string]string{}

	if err := unmarshal(&bz); err != nil {
		return annotations, err
	}

	for _, ann := range bz.Annotations {
		annotations[ann.Name] = ann.Value
	}
	return annotations, nil
}

func unmarshalPorts(unmarshal func(interface{}) error) ([]int, error) {
	var portYAML struct {
		Port  string `yaml:"port,omitempty"`
		Ports string `yaml:"ports,omitempty"`
	}

	var ports []int

	if err := unmarshal(&portYAML); err != nil {
		return ports, err
	}

	if portYAML.Ports != "" {
		ports = stringToIntArray(portYAML.Ports)
	} else if portYAML.Port != "" {
		ports = stringToIntArray(portYAML.Port)
	} else {
		ports = []int{80}
	}
	return ports, nil
}

func stringToIntArray(str string) []int {
	var retval []int

	pstr := strings.Split(str, ",")
	for _, p := range pstr {
		j, err := strconv.Atoi(p)
		if err == nil {
			retval = append(retval, j)
		}
	}
	return retval
}
