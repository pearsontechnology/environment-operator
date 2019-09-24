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
	Version         string                        `json:"version"`
	Options         map[string]interface{}        `json:"options"`
	Replicas        int                           `json:"replicas,omitempty"`
	TargetNamespace string                        `json:"target_namespace,omitempty"`
	Chart           string                        `json:"chart,omitempty"`
	Repo            string                        `json:"repo,omitempty"`
	Set             map[string]intstr.IntOrString `json:"set,omitempty"`
	ValuesContent   string                        `json:"values_content,omitempty"`
	Ignore          bool                          `json:"ignore,omitempty"`
}

// PrsnExternalResourceList is a list of PrsnExternalResource
type PrsnExternalResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PrsnExternalResource `json:"items"`
}

// DeepCopyObject required to satisfy Object interface
func (tpr PrsnExternalResource) DeepCopyObject() runtime.Object {
	return new(PrsnExternalResource)
}

// DeepCopyObject required to satisfy Object interface
func (tpr PrsnExternalResourceList) DeepCopyObject() runtime.Object {
	return new(PrsnExternalResource)
}
