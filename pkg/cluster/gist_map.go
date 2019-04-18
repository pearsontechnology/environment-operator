package cluster

import (
	"sort"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	v1batch "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// GistMap holds a list of bitesize.Gist objects, representing the
// bitesize config gist path
type GistMap map[string]*bitesize.Gist

// AddConfigMap adds imported ConfigMap resource to GistMap
func (m GistMap) AddConfigMap(gist v1.ConfigMap) *bitesize.Gist {
	name := gist.Name
	// Create with some defaults -- defaults should probably live in bitesize.Gist
	if m[name] == nil {
		m[name] = &bitesize.Gist{
			Name:      name,
			Type:      bitesize.TypeConfigMap,
			ConfigMap: gist,
		}
	}
	return m[name]
}

// AddJob adds imported v1batch.Job Gist to GistMap
func (m GistMap) AddJob(gist v1batch.Job) *bitesize.Gist {
	if m[gist.Name] == nil {
		m[gist.Name] = &bitesize.Gist{
			Name: gist.Name,
			Type: bitesize.TypeJob,
			Job:  gist,
		}
	}
	return m[gist.Name]
}

// AddCronJob adds imported v1beta1.CronJob Gist to GistMap
func (m GistMap) AddCronJob(gist v1beta1.CronJob) *bitesize.Gist {
	if m[gist.Name] == nil {
		m[gist.Name] = &bitesize.Gist{
			Name:    gist.Name,
			Type:    bitesize.TypeCronJob,
			CronJob: gist,
		}
	}
	return m[gist.Name]
}

// Gists extracts a sorted list of bitesize.Gist type out from
// ImportMap type
func (m GistMap) Gists() bitesize.Gists {
	var gists bitesize.Gists

	for _, v := range m {
		gists = append(gists, *v)
	}

	sort.Sort(gists)
	return gists
}
