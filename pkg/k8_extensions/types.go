package k8_extensions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SupportedCustomResources contains all supported CRDs on bitesize cluster.
var SupportedCustomResources = []string{
	"mongo", "mysql", "cassandra", "redis", "zookeeper", "kafka", "postgres", "neptune", "sns", "msk", "docdb", "cb", "sqs", "s3", "helmchart",
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
}

// PrsnExternalResourceList is a list of PrsnExternalResource
type PrsnExternalResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PrsnExternalResource `json:"items"`
}

// Server represents format for these mappings
type Server struct {
	Port  *Port    `json:"port,omitempty"`
	Bind  string   `json:"bind,omitempty"`
	Hosts []string `json:"hosts,omitempty"`
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
