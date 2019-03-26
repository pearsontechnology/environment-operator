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
	path := res.Labels["path"]
	// Create with some defaults -- defaults should probably live in bitesize.Resource
	if m[path] == nil {
		m[path] = &bitesize.Resource{
			Path:      path,
			Type:      "configmap",
			ConfigMap: res,
		}
	}
	return m[path]
}

// AddJob adds imported v1batch.Job resource to resourcemap
func (m ResourceMap) AddJob(res v1batch.Job) *bitesize.Resource {
	path := res.Labels["path"]
	// Create with some defaults -- defaults should probably live in bitesize.Resource
	if m[path] == nil {
		m[path] = &bitesize.Resource{
			Path: path,
			Type: "job",
			Job:  res,
		}
	}
	return m[path]
}

// AddCronJob adds imported v1beta1.CronJob resource to resourcemap
func (m ResourceMap) AddCronJob(res v1beta1.CronJob) *bitesize.Resource {
	path := res.Labels["path"]
	// Create with some defaults -- defaults should probably live in bitesize.Resource
	if m[path] == nil {
		m[path] = &bitesize.Resource{
			Path:    path,
			Type:    "cronjob",
			CronJob: res,
		}
	}
	return m[path]
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
