package bitesize

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/config"
	validator "gopkg.in/validator.v2"
)

// Service represents a single service and it's configuration,
// running in environment
type Service struct {
	Name            string                  `yaml:"name" validate:"nonzero"`
	ExternalURL     []string                `yaml:"-"`
	Backend         string                  `yaml:"backend"`
	BackendPort     int                     `yaml:"backend_port"`
	Ports           []int                   `yaml:"-"` // Ports have custom unmarshaler
	Ssl             string                  `yaml:"ssl,omitempty" validate:"regexp=^(true|false)*$"`
	Version         string                  `yaml:"version,omitempty"`
	Application     string                  `yaml:"application,omitempty"`
	Replicas        *int32                  `yaml:"replicas,omitempty"`
	Deployment      *DeploymentSettings     `yaml:"deployment,omitempty"`
	HPA             HorizontalPodAutoscaler `yaml:"hpa" validate:"hpa"`
	Requests        ContainerRequests       `yaml:"requests" validate:"requests"`
	Limits          ContainerLimits         `yaml:"limits" validate:"limits"`
	HealthCheck     *HealthCheck            `yaml:"health_check,omitempty"`
	LivenessProbe   *Probe                  `yaml:"liveness_probe,omitempty"`
	ReadinessProbe  *Probe                  `yaml:"readiness_probe,omitempty"`
	EnvVars         []EnvVar                `yaml:"env,omitempty"`
	Commands        []string                `yaml:"command,omitempty"`
	Annotations     map[string]string       `yaml:"-"` // Annotations have custom unmarshaler
	Volumes         []Volume                `yaml:"volumes,omitempty"`
	Options         map[string]interface{}  `yaml:"-"` // Options have custom unmarshaler
	HTTP2           string                  `yaml:"http2,omitempty" validate:"regexp=^(true|false)*$"`
	HTTPSOnly       string                  `yaml:"httpsOnly,omitempty" validate:"regexp=^(true|false)*$"`
	HTTPSBackend    string                  `yaml:"httpsBackend,omitempty" validate:"regexp=^(true|false)*$"`
	Type            string                  `yaml:"type,omitempty"`
	Status          ServiceStatus           `yaml:"status,omitempty"`
	DatabaseType    string                  `yaml:"database_type,omitempty" validate:"regexp=^(mongo)*$"`
	GracePeriod     *int64                  `yaml:"graceperiod,omitempty"`
	ResourceVersion string                  `yaml:"resourceVersion,omitempty"`
	// XXX          map[string]interface{} `yaml:",inline"`
}

// ServiceStatus represents cluster service's status metrics
type ServiceStatus struct {
	DeployedAt        string
	AvailableReplicas int
	DesiredReplicas   int
	CurrentReplicas   int
}

// Services implement sort.Interface
type Services []Service

// Default Replica count
var DefaultReplicas int32 = 1

// ServiceWithDefaults returns new *Service object with default values set
func ServiceWithDefaults() *Service {
	return &Service{
		Ports:    []int{80},
		Replicas: &DefaultReplicas,
		Limits: ContainerLimits{
			Memory: config.Env.LimitDefaultMemory,
			CPU:    config.Env.LimitDefaultCPU,
		},
		Requests: ContainerRequests{
			CPU: config.Env.RequestsDefaultCPU,
		},
		Deployment:  &DeploymentSettings{},
		ExternalURL: []string{},
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

	externalURL, err := unmarshalExternalURL(unmarshal)
	if err != nil {
		return fmt.Errorf("service.external_url.%s", err.Error())
	}

	unmarshalOptions, err := unmarshalOptions(unmarshal)
	if err != nil {
		return fmt.Errorf("service.options.%s", err.Error())
	}

	type plain Service
	if err = unmarshal((*plain)(ee)); err != nil {
		return fmt.Errorf("service.%s", err.Error())
	}

	*e = *ee
	e.Ports = ports
	e.Annotations = annotations
	e.ExternalURL = externalURL
	e.Options = unmarshalOptions

	if e.Type != "" {
		e.Ports = nil
	}

	// annotation := Annotation{Name: "Name", Value: e.Name}
	// e.Annotations = append(e.Annotations, annotation)

	if e.HPA.MinReplicas != 0 {
		e.Replicas = nil
	}

	if e.HPA.MinReplicas != 0 && e.HPA.Metric.Name == "" {
		e.HPA.Metric = Metric{Name: "cpu", TargetAverageUtilization: int32(80)}
	}

	if err = validator.Validate(e); err != nil {
		return fmt.Errorf("service.%s", err.Error())
	}

	return nil
}

// HasExternalURL checks if the service has an external_url defined
func (e Service) HasExternalURL() bool {
	return len(e.ExternalURL) != 0
}

// IsBlueGreenParentDeployment verifies if deployment method set for the service
// is bluegreen
func (e Service) IsBlueGreenParentDeployment() bool {
	if e.Deployment == nil {
		return false
	}
	return e.DeploymentMethod() == "bluegreen"
}

// IsBlueGreenChildDeployment returns true if this service is a child of main bluegreen service
func (e Service) IsBlueGreenChildDeployment() bool {
	if e.Deployment == nil || e.Deployment.BlueGreen == nil {
		return false
	}
	if e.Deployment.BlueGreen.DeploymentColour != nil {
		return true
	}
	return false
}

// DeploymentMethod returns deployment method for service. rolling-upgrade or bluegreen
func (e Service) DeploymentMethod() string {
	if e.Deployment == nil {
		return "rolling-upgrade"
	}
	return e.Deployment.Method
}

// InactiveDeploymentTag returns inactive deployment in bluegreen set
func (e Service) InactiveDeploymentTag() BlueGreenServiceSet {
	if e.Deployment == nil || e.Deployment.BlueGreen == nil {
		return BlueService
	}
	if *e.Deployment.BlueGreen.Active == BlueService {
		return GreenService
	}
	return BlueService
}

// ActiveDeploymentName returns a fully formatted name for active bluegreen deployment
func (e Service) ActiveDeploymentName() string {
	if !e.IsBlueGreenParentDeployment() {
		return e.Name
	}
	return fmt.Sprintf("%s-%s", e.Name, e.ActiveDeploymentTag().String())
}

// InactiveDeploymentName returns a fully formatted name for the inactive bluegreen deployment
func (e Service) InactiveDeploymentName() string {
	if !e.IsBlueGreenParentDeployment() {
		return ""
	}
	return fmt.Sprintf("%s-%s", e.Name, e.InactiveDeploymentTag().String())
}

// IsActiveBlueGreenDeployment returns a boolean specifying whether current "child" service
// is service active traffic
func (e Service) IsActiveBlueGreenDeployment() bool {
	if e.Deployment == nil || e.Deployment.BlueGreen == nil {
		return false
	}
	return e.Deployment.BlueGreen.ActiveFlag
}

// ActiveDeploymentTag returns active deployment in bluegreen set
func (e Service) ActiveDeploymentTag() BlueGreenServiceSet {
	if e.Deployment == nil || e.Deployment.BlueGreen == nil {
		return 0
	}
	return *e.Deployment.BlueGreen.Active
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

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func unmarshalOptions(unmarshal func(interface{}) error) (map[string]interface{}, error) {
	var bz struct {
		Options map[string]interface{} `yaml:"options,omitempty"`
	}

	options := map[string]interface{}{}

	if err := unmarshal(&bz); err != nil {
		return options, err
	}

	for k, v := range bz.Options {
		options[k] = cleanupMapValue(v)
	}

	if len(options) == 0 {
		return nil, nil
	}

	return options, nil
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

func unmarshalExternalURL(unmarshal func(interface{}) error) ([]string, error) {

	var u struct {
		URL interface{} `yaml:"external_url,omitempty"`
	}
	urls := []string{}

	if err := unmarshal(&u); err != nil {
		return urls, err
	}

	switch v := u.URL.(type) {
	case string:
		urls = append(urls, v)
	case []interface{}:
		for _, url := range v {
			urls = append(urls, reflect.ValueOf(url).String())
		}
	case nil:
		return urls, nil
	default:
		return nil, fmt.Errorf("unsupported type %v declared for external_url %v", v, u)
	}

	return urls, nil
}

func (v *Volume) UnmarshalYAML(unmarshal func(interface{}) error) error {
	vv := &Volume{
		Modes:        "ReadWriteOnce",
		provisioning: "dynamic",
		Type:         "ebs",
	}

	type plain Volume
	if err := unmarshal((*plain)(vv)); err != nil {
		return fmt.Errorf("volume.%s", err.Error())
	}

	*v = *vv
	return nil
}

func (v *Volume) HasManualProvisioning() bool {
	if v.provisioning == "manual" {
		return true
	}
	return false
}

// IsSecretVolume returns true if volume is secret
func (v *Volume) IsSecretVolume() bool {
	if strings.ToLower(v.Type) == "secret" {
		return true
	}
	return false
}
