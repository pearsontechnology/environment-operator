package k8_extensions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SupportedCustomResources contains all supported CRDs on bitesize cluster.
var SupportedCustomResources = []string{
	"aurora", "mongo", "mysql", "cassandra", "redis", "zookeeper", "kafka",
	"postgres", "neptune", "sns", "msk", "docdb", "cb", "sqs", "s3",
	"helmchart", "dynamodb", "serviceentry", "gateway", "virtualservice",
}

// SupportedCustomResourceAPIVersions contains all supported CRD API versions on bitesize cluster.
var SupportedCustomResourceAPIVersions = []string{
	"prsn.io/v1", "helm.kubedex.com/v1", "networking.istio.io/v1alpha3",
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
	Addresses       []string                      `json:"addresses,omitempty"`
	Ports           []*Port                       `json:"ports,omitempty"`
	Location        string                        `json:"location,omitempty"`
	Resolution      string                        `json:"resolution,omitempty"`
	Endpoints       []*ServiceEntry_Endpoint      `json:"endpoints,omitempty"`
	ExportTo        []string                      `json:"export_to,omitempty"`
	SubjectAltNames []string                      `json:"subject_alt_names,omitempty"`
	HTTP            []*HTTPRoute                  `json:"http,omitempty"`
}

type ServiceEntry_Endpoint struct {
	Address  string            `json:"address,omitempty"`
	Ports    map[string]uint32 `json:"ports,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Network  string            `json:"network,omitempty"`
	Locality string            `json:"locality,omitempty"`
	Weight   uint32            `json:"weight,omitempty"`
}

type ExternalSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ExternalSecret `json:"items"`
}

type ExternalSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	SecretDescriptor  ExternalSecretSecretDescriptor `json:"secretDescriptor"`
}

type ExternalSecretSecretDescriptor struct {
	// RoleArn     string              `json:"roleArn, omitempty"`
	// VaultRole   string              `json:"vaultRole, omitempty"`
	BackendType string              `json:"backendType, omitempty"`
	Type        string              `json:"type, omitempty"`
	Compressed  bool                `json:"compressed, omitempty"`
	Data        []map[string]string `json:"data, omitempty"`
	// DataFrom    []string            `json:"dataFrom, omitempty"`
}

// HTTPRoute represents format for these mappings
type HTTPRoute struct {
	Name  string                  `json:"name,omitempty"`
	Match []*HTTPMatchRequest     `json:"match,omitempty"`
	Route []*HTTPRouteDestination `json:"route,omitempty"`
}

// HTTPMatchRequest represents format for these mappings
type HTTPMatchRequest struct {
	Name string        `json:"name,omitempty"`
	URI  *StringPrefix `json:"uri,omitempty"`
}

// StringPrefix represents format for these mappings
type StringPrefix struct {
	Prefix string `json:"prefix,omitempty"`
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

func (es ExternalSecret) DeepCopyObject() runtime.Object {
	return new(ExternalSecret)
}

func (es ExternalSecretList) DeepCopyObject() runtime.Object {
	return new(ExternalSecretList)
}
