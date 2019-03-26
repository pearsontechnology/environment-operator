package bitesize

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
	v1batch "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// Resource represent a resource
type Resource struct {
	Path      string          `yaml:"path"`
	Type      string          `yaml:"configamp"`
	ConfigMap v1.ConfigMap    `yaml:"-"`
	Job       v1batch.Job     `yaml:"-"`
	CronJob   v1beta1.CronJob `yaml:"-"`
}

// Imports is the struct to hold all imported resources per env
type Imports []Resource

func (slice Imports) Len() int {
	return len(slice)
}

func (slice Imports) Less(i, j int) bool {
	return slice[i].Path < slice[j].Path
}

func (slice Imports) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// LoadResource loads named resource from a filename with a given path
func LoadResource(path string, rstype string) (*Resource, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := &Resource{}

	switch rstype {
	case "configmap":
		{
			if err := yaml.Unmarshal(contents, t.ConfigMap); err != nil {
				return nil, err
			}

			labels := t.ConfigMap.GetLabels()
			labels["creator"] = "pipeline"
			t.ConfigMap.SetLabels(labels)
		}
	case "job":
		{
			if err := yaml.Unmarshal(contents, t.Job); err != nil {
				return nil, err
			}

			labels := t.ConfigMap.GetLabels()
			labels["creator"] = "pipeline"
			t.Job.SetLabels(labels)
		}
	case "cronjob":
		{
			if err := yaml.Unmarshal(contents, t.CronJob); err != nil {
				return nil, err
			}

			labels := t.ConfigMap.GetLabels()
			labels["creator"] = "pipeline"
			t.CronJob.SetLabels(labels)
		}
	}

	return t, nil
}

// Find returns service with a name match
func (slice Imports) Find(path string, rstype string) *Resource {
	for _, s := range slice {
		if s.Path == path && s.Type == rstype {
			return &s
		}
	}
	return nil
}

// FindConfigMapByName returns service with a name match
func (slice Imports) FindConfigMapByName(name string) *Resource {
	for _, s := range slice {
		if s.Type == "configmap" && s.ConfigMap.Name == name {
			return &s
		}
	}
	return nil
}
