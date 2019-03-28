package bitesize

import (
	"os"
	"path"

	"github.com/pearsontechnology/environment-operator/pkg/util"
	v1batch "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Resource represent a resource
type Resource struct {
	Name       string          `yaml:"name,omitempty"`
	Path       string          `yaml:"path,omitempty"`
	Files      []string        `yaml:"files,omitempty"`
	AppendHash bool            `yaml:"append_hash"`
	Type       string          `yaml:"type"`
	ConfigMap  v1.ConfigMap    `yaml:"-"`
	Job        v1batch.Job     `yaml:"-"`
	CronJob    v1beta1.CronJob `yaml:"-"`
}

// ImportsRepository contains the repository info all the imports per env
type ImportsRepository struct {
	Remote string `yaml:"remote"`
	Branch string `yaml:"branch,omitempty" default:"master"`
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
func LoadResource(res Resource, localPath string) (*Resource, error) {
	t := &Resource{}

	if len(res.Path) > 0 {
		resPath := path.Join(localPath, res.Path)

		file, err := os.Open(resPath)
		if err != nil {
			return nil, err
		}
		decoder := yaml.NewYAMLToJSONDecoder(file)

		switch res.Type {
		case "configmap":
			{

				if err := decoder.Decode(&t.ConfigMap); err != nil {
					return nil, err
				}
				labels := t.ConfigMap.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				t.ConfigMap.SetLabels(labels)
			}
		case "job":
			{
				if err := decoder.Decode(&t.Job); err != nil {
					return nil, err
				}
				labels := t.Job.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				t.Job.SetLabels(labels)
			}
		case "cronjob":
			{
				if err := decoder.Decode(&t.CronJob); err != nil {
					return nil, err
				}
				labels := t.CronJob.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				t.CronJob.SetLabels(labels)
			}
		}
	}

	if len(t.Files) > 0 {
		// set absolute paths
		for k, v := range t.Files {
			t.Files[k] = path.Join(localPath, v)
		}
		// if the configmap is to be generated from files
		generator := util.ConfigMapGenerator{
			Name:        res.Name,
			FileSources: res.Files,
			AppendHash:  res.AppendHash,
		}
		cfmap, err := generator.Generate()
		if err != nil {
			return nil, err
		}
		t.ConfigMap = *cfmap

		labels := t.ConfigMap.GetLabels()
		labels["creator"] = "pipeline"
		t.ConfigMap.SetLabels(labels)
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
