package bitesize

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	v1batch "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// TypeConfigMap k8s configmap type
	TypeConfigMap string = "configmap"
	// TypeJob k8s job type
	TypeJob string = "job"
	// TypeCronJob k8s cronjob type
	TypeCronJob string = "cronjob"
	// TypeSecret k8s secret type
	TypeSecret string = "secret"
)

// Gist represent a resource
type Gist struct {
	Name      string          `yaml:"name"`
	Path      string          `yaml:"path,omitempty"`
	Files     []string        `yaml:"files,omitempty"`
	Type      string          `yaml:"type"`
	ConfigMap v1.ConfigMap    `yaml:"-"`
	Job       v1batch.Job     `yaml:"-"`
	CronJob   v1beta1.CronJob `yaml:"-"`
	Secret    v1.Secret       `yaml:"-"`
}

// GistsRepository contains the repository info all the imports per env
type GistsRepository struct {
	Remote string `yaml:"remote"`
	Branch string `yaml:"branch,omitempty" default:"master"`
}

// Gists is the struct to hold all imported resources per env
type Gists []Gist

func (slice Gists) Len() int {
	return len(slice)
}

func (slice Gists) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice Gists) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// LoadResource loads named resource from a filename with a given path
func LoadResource(res *Gist, namespace, localPath string) error {
	if len(res.Path) > 0 {
		resPath := path.Join(localPath, res.Path)

		file, err := os.Open(resPath)
		if err != nil {
			return err
		}

		decoder := k8Yaml.NewYAMLOrJSONDecoder(file, 1000)

		switch res.Type {
		case TypeConfigMap:
			{
				log.Debugf("adding ConfigMap %s from ConfigMap path %s", res.Name, res.Path)
				if err := decoder.Decode(&res.ConfigMap); err != nil {
					return err
				}
				labels := res.ConfigMap.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				res.ConfigMap.ObjectMeta.SetLabels(labels)
				res.ConfigMap.ObjectMeta.SetName(res.Name)
				// override metadata namespace to current environment namespace
				res.ConfigMap.ObjectMeta.SetNamespace(namespace)
			}
		case TypeJob:
			{
				if err := decoder.Decode(&res.Job); err != nil {
					return err
				}
				labels := res.Job.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				res.Job.SetLabels(labels)
				res.ConfigMap.ObjectMeta.SetLabels(labels)
				res.ConfigMap.ObjectMeta.SetName(res.Name)
				// override metadata namespace to current environment namespace
				res.ConfigMap.ObjectMeta.SetNamespace(namespace)
			}
		case TypeCronJob:
			{
				if err := decoder.Decode(&res.CronJob); err != nil {
					return err
				}
				labels := res.CronJob.GetLabels()
				if labels == nil {
					labels = map[string]string{
						"creator": "pipeline",
					}
				} else {
					labels["creator"] = "pipeline"
				}
				res.ConfigMap.ObjectMeta.SetLabels(labels)
				res.ConfigMap.ObjectMeta.SetName(res.Name)
				// override metadata namespace to current environment namespace
				res.ConfigMap.ObjectMeta.SetNamespace(namespace)
			}
		}
	}

	if len(res.Files) > 0 {
		log.Debugf("generating configmap %s from file sources", res.Name)
		// set absolute paths
		for k, v := range res.Files {
			res.Files[k] = path.Join(localPath, v)
			log.Debugf("file: %s", res.Files[k])
		}
		// if the ConfigMap is to be generated from files
		generator := util.ConfigMapGenerator{
			Name:        res.Name,
			FileSources: res.Files,
			AppendHash:  false,
		}
		cfmap, err := generator.Generate()
		if err != nil {
			return err
		}

		res.ConfigMap = *cfmap

		labels := res.ConfigMap.GetLabels()
		if labels == nil {
			labels = map[string]string{
				"creator": "pipeline",
			}
		} else {
			labels["creator"] = "pipeline"
		}
		res.ConfigMap.ObjectMeta.SetLabels(labels)
		res.ConfigMap.ObjectMeta.SetName(res.Name)
		// override metadata namespace to current environment namespace
		res.ConfigMap.ObjectMeta.SetNamespace(namespace)
	}

	return nil
}

// Find returns service with a name match
// path is a UNC path relative to the configured repository root
// rstype is the resource type of the imported resource
//   available type are:
//      - configmap
//      - job
//      - cronjob
// if resource found returns the resource else returns nil
func (slice Gists) Find(path string, rstype string) *Gist {
	for _, s := range slice {
		if s.Path == path && s.Type == rstype {
			return &s
		}
	}
	return nil
}

// FindByName returns configmap resource matched with name parameter
// rstype is the resource type of the imported resource
//   available type are:
//      - bitesize.TypeConfigMap
//      - bitesize.TypeJob
//      - bitesize.TypeCronJob
// if resource found returns the resource else returns nil
func (slice Gists) FindByName(name string, rstype string) *Gist {
	for _, s := range slice {
		if s.Type == rstype && s.Name == name {
			return &s
		}
	}
	return nil
}

// FindByType returns slice of resources matched with type parameter
// rstype is the resource type of the imported resource
//   available type are:
//      - bitesize.TypeConfigMap
//      - bitesize.TypeJob
//      - bitesize.TypeCronJob
// if resource found returns the resource else returns nil
func (slice Gists) FindByType(rstype string) Gists {
	res := Gists{}
	for _, s := range slice {
		if s.Type == rstype {
			res = append(res, s)
		}
	}
	return res
}
