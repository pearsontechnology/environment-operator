package bitesize

import (
	"io/ioutil"

	"github.com/pearsontechnology/environment-operator/pkg/util"
	yaml "gopkg.in/yaml.v2"
	v1batch "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// Resource represent a resource
type Resource struct {
	Name       string          `yaml:"name,omitempty"`
	File       string          `yaml:"file,omitempty"`
	Files      []string        `yaml:"files,omitempty"`
	AppendHash bool            `yaml:"append_hash"`
	Type       string          `yaml:"type"`
	ConfigMap  v1.ConfigMap    `yaml:"-"`
	Job        v1batch.Job     `yaml:"-"`
	CronJob    v1beta1.CronJob `yaml:"-"`
}

// Imports is the struct to hold all imported resources per env
type Imports []Resource

func (slice Imports) Len() int {
	return len(slice)
}

func (slice Imports) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice Imports) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// LoadResource loads named resource from a filename with a given path
func LoadResource(res Resource) (*Resource, error) {
	t := &Resource{}

	switch res.Type {
	case "configmap":
		{
			generator := util.ConfigMapGenerator{
				Name:        res.Name,
				FileSources: res.Files,
				AppendHash:  res.AppendHash,
			}

			cfmap, err := generator.Generate()
			if err != nil {
				return nil, err
			}
			labels := cfmap.GetLabels()
			labels["creator"] = "pipeline"
			cfmap.SetLabels(labels)

			t.ConfigMap = *cfmap
		}
	case "job":
		{
			contents, err := ioutil.ReadFile(res.File)
			if err != nil {
				return nil, err
			}

			if err := yaml.Unmarshal(contents, t.Job); err != nil {
				return nil, err
			}

			labels := t.ConfigMap.GetLabels()
			labels["creator"] = "pipeline"
			t.Job.SetLabels(labels)
		}
	case "cronjob":
		{
			contents, err := ioutil.ReadFile(res.File)
			if err != nil {
				return nil, err
			}

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
func (slice Imports) Find(file string, rstype string) *Resource {
	for _, s := range slice {
		if s.File == file && s.Type == rstype {
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
