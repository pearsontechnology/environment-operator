package cluster

import (
	"sort"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	v1batch "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// ResourceMap holds a list of bitesize.Resource objects, representing the
// bitesize config import path
type ResourceMap map[string]*bitesize.Resource

// AddConfigMap adds imported ConfigMap resource to resourcemap
func (m ResourceMap) AddConfigMap(res v1.ConfigMap) *bitesize.Resource {
	name := res.Name
	// Create with some defaults -- defaults should probably live in bitesize.Resource
	if m[name] == nil {
		m[name] = &bitesize.Resource{
			Name:      name,
			Type:      bitesize.TypeConfigMap,
			ConfigMap: res,
		}
	}
	return m[name]
}

// AddJob adds imported v1batch.Job resource to resourcemap
func (m ResourceMap) AddJob(res v1batch.Job) *bitesize.Resource {
	if m[res.Name] == nil {
		m[res.Name] = &bitesize.Resource{
			Name: res.Name,
			Type: bitesize.TypeJob,
			Job:  res,
		}
	}
	return m[res.Name]
}

// AddCronJob adds imported v1beta1.CronJob resource to resourcemap
func (m ResourceMap) AddCronJob(res v1beta1.CronJob) *bitesize.Resource {
	if m[res.Name] == nil {
		m[res.Name] = &bitesize.Resource{
			Name:    res.Name,
			Type:    bitesize.TypeCronJob,
			CronJob: res,
		}
	}
	return m[res.Name]
}

// Resources extracts a sorted list of bitesize.Import type out from
// ImportMap type
func (m ResourceMap) Resources() bitesize.Imports {
	var resourceList bitesize.Imports

	for _, v := range m {
		resourceList = append(resourceList, *v)
	}

	sort.Sort(resourceList)
	return resourceList
}
