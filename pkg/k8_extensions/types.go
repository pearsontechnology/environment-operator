package k8_extensions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SupportedCustomResources contains all supported CRDs on bitesize cluster.
var SupportedCustomResources = []string{
	"mongo", "mysql", "cassandra", "redis", "zookeeper", "kafka", "postgres", "neptune", "sns", "msk", "docdb", "cb", "sqs", "s3", "helmchart", "dynamodb",
}

// SupportedCustomResourceAPIVersions contains all supported CRD API versions on bitesize cluster.
var SupportedCustomResourceAPIVersions = []string{
	"prsn.io/v1", "helm.kubedex.com/v1",
}

// PrsnExternalResource represents CustomResources mapped from
// kubernetes to externally running services (e.g. RDS, cassandra, mongo etc.)
type PrsnExternalResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec PrsnExternalResourceSpec `json:"spec"`
}

// PrsnExternalResourceSpec represents format for these mappings - which is
// basically it's version and  options
type PrsnExternalResourceSpec struct {
	Version         string                        `json:"version,omitempty"`
	Options         map[string]interface{}        `json:"options,omitempty"`
	Replicas        int                           `json:"replicas,omitempty"`
	TargetNamespace string                        `json:"targetNamespace,omitempty"`
	Chart           string                        `json:"chart,omitempty"`
	Repo            string                        `json:"repo,omitempty"`
	Set             map[string]intstr.IntOrString `json:"set,omitempty"`
	ValuesContent   string                        `json:"valuesContent,omitempty"`
	Ignore          bool                          `json:"ignore,omitempty"`
	Selector        map[string]string             `json:"selector,omitempty"`
	Servers         []*Server                     `json:"servers,omitempty"`
	Gateways        []string                      `json:"gateways,omitempty"`
	Hosts           []string                      `json:"hosts,omitempty"`
	HTTP            []*HTTPRoute                  `json:"http,omitempty"`
}

// HTTPRoute represents format for these mappings
type HTTPRoute struct {
	Name  string                  `json:"name,omitempty"`
	Match []*HTTPMatchRequest     `json:"match,omitempty"`
	Route []*HTTPRouteDestination `json:"route,omitempty"`
}

// HTTPMatchRequest represents format for these mappings
type HTTPMatchRequest struct {
	Name string       `json:"name,omitempty"`
	URI  *StringExact `json:"uri,omitempty"`
}

// StringExact represents format for these mappings
type StringExact struct {
	Exact string `json:"exact,omitempty"`
}

// HTTPRouteDestination represents format for these mappings
type HTTPRouteDestination struct {
	Destination *Destination `json:"destination,omitempty"`
}

// Destination represents format for these mappings
type Destination struct {
	Host   string        `json:"host,omitempty"`
	Subset string        `json:"subset,omitempty"`
	Port   *PortSelector `json:"port,omitempty"`
}

// PortSelector represents format for these mappings
type PortSelector struct {
	Number uint32 `json:"number,omitempty"`
}

// PrsnExternalResourceList is a list of PrsnExternalResource
type PrsnExternalResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PrsnExternalResource `json:"items"`
}

// Server represents format for these mappings
type Server struct {
	Port  *Port             `json:"port,omitempty"`
	Bind  string            `json:"bind,omitempty"`
	Hosts []string          `json:"hosts,omitempty"`
	TLS   *ServerTLSOptions `json:"tls,omitempty"`
}

// ServerTLSOptions is set of TLS related options that govern the server's behavior
type ServerTLSOptions struct {
	Mode           string `json:"mode,omitempty"`
	CredentialName string `json:"credential_name,omitempty"`
}

// Port represents format for these mappings
type Port struct {
	Number   uint32 `json:"number,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Name     string `json:"name,omitempty"`
}

// DeepCopyObject required to satisfy Object interface
func (tpr PrsnExternalResource) DeepCopyObject() runtime.Object {
	return new(PrsnExternalResource)
}

// DeepCopyObject required to satisfy Object interface
func (tpr PrsnExternalResourceList) DeepCopyObject() runtime.Object {
	return new(PrsnExternalResource)
}
