package k8_extensions

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

// SupportedThirdPartyResources contains all supported TPRs on bitesize
// cluster.
var SupportedThirdPartyResources = []string{
	"mongo", "mysql", "cassandra", "redis",
}

// PrsnExternalResource represents ThirdpartyResources mapped from
// kubernetes to externally running services (e.g. RDS, cassandra, mongo etc.)
type PrsnExternalResource struct {
	unversioned.TypeMeta `json:", inline"`

	ObjectMeta v1.ObjectMeta `json:"metadata"`

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
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta           unversioned.ListMeta `json:"metadata"`

	Items []PrsnExternalResource `json:"items"`
}

// GetObjectKind required to satisfy Object interface
func (tpr *PrsnExternalResource) GetObjectKind() unversioned.ObjectKind {
	return &tpr.TypeMeta
}

// GetObjectMeta required to satisfy ObjectMetaAccessor interface
func (tpr *PrsnExternalResource) GetObjectMeta() unversioned.ObjectKind {
	return &tpr.TypeMeta
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
