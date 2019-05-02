package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMap type actions on configmaps in k8s cluster
type ConfigMap struct {
	kubernetes.Interface
	Namespace string
}

// Get returns ingress object from the k8s by name
func (client *ConfigMap) Get(name string) (*v1.ConfigMap, error) {
	return client.
		CoreV1().
		ConfigMaps(client.Namespace).
		Get(name, getOptions())
}

// Exist returns boolean value if ingress exists in k8s
func (client *ConfigMap) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates ingress in k8s
func (client *ConfigMap) Apply(resource *v1.ConfigMap) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)

}

// Update updates existing ingress in k8s
func (client *ConfigMap) Update(resource *v1.ConfigMap) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		CoreV1().
		ConfigMaps(client.Namespace).
		Update(resource)
	return err
}

// Create creates new configmap in k8s
func (client *ConfigMap) Create(resource *v1.ConfigMap) error {
	_, err := client.
		CoreV1().
		ConfigMaps(client.Namespace).
		Create(resource)
	return err
}

// Destroy deletes configmap from the k8 cluster
func (client *ConfigMap) Destroy(name string) error {
	return client.CoreV1().ConfigMaps(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *ConfigMap) List() ([]v1.ConfigMap, error) {
	list, err := client.CoreV1().ConfigMaps(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
