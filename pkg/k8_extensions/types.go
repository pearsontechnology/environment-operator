package k8_extensions

import (
	"encoding/json"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SupportedThirdPartyResources contains all supported TPRs on bitesize
// cluster.
var SupportedThirdPartyResources = []string{
	"mongo", "mysql", "cassandra", "redis",
}

// PrsnExternalResource represents ThirdpartyResources mapped from
// kubernetes to externally running services (e.g. RDS, cassandra, mongo etc.)
type PrsnExternalResource struct {
	meta.TypeMeta `json:", inline"`

	ObjectMeta meta.ObjectMeta `json:"metadata"`

	Spec PrsnExternalResourceSpec `json:"spec"`
}

// PrsnExternalResourceSpec represents format for these mappings - which is
// basically it's version and  options
type PrsnExternalResourceSpec struct {
	Version  string            `json:"version"`
	Options  map[string]string `json:"options"`
	Replicas int               `json:"replicas"`
}

// PrsnExternalResourceList is a list of PrsnExternalResource
type PrsnExternalResourceList struct {
	meta.TypeMeta `json:",inline"`
	ObjectMeta    meta.ListMeta `json:"metadata"`

	Items []PrsnExternalResource `json:"items"`
}

// GetObjectKind required to satisfy Object interface
func (tpr PrsnExternalResource) GetObjectKind() schema.ObjectKind {
	return &tpr.TypeMeta
}

// GetObjectMeta required to satisfy ObjectMetaAccessor interface
// func (tpr PrsnExternalResource) GetObjectMeta() v1.ObjectMeta {
// 	return tpr.ObjectMeta
// }

func (tpr PrsnExternalResource) GetObjectMeta() meta.Object {

	return &tpr.ObjectMeta
}

// GetObjectKind required to satisfy Object interface
// func (tprList *PrsnExternalResourceList) GetObjectKind() unversioned.ObjectKind {
// 	return &tprList.TypeMeta
// }

// GetListMeta required to satisfy ListMetaAccessor interface
// func (tprList *PrsnExternalResourceList) GetListMeta() v1.List {
// 	return &tprList.ObjectMeta
// }

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type prsnExternalResourceListCopy PrsnExternalResourceList
type prsnExternalResourceCopy PrsnExternalResource

// UnmarshalJSON unmarshals JSON into PrsnExternalResource
func (tpr *PrsnExternalResource) UnmarshalJSON(data []byte) error {
	tmp := prsnExternalResourceCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := PrsnExternalResource(tmp)
	*tpr = tmp2
	return nil
}

// UnmarshalJSON unmarshals JSON into PrsnExternalResourceList
func (tprList *PrsnExternalResourceList) UnmarshalJSON(data []byte) error {
	tmp := prsnExternalResourceListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := PrsnExternalResourceList(tmp)
	*tprList = tmp2
	return nil
}
