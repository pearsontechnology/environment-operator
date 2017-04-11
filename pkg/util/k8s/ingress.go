package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Ingress type actions on ingresses in k8s cluster
type Ingress struct {
	kubernetes.Interface
	Namespace string
}

// Get returns ingress object from the k8s by name
func (client *Ingress) Get(name string) (*v1beta1.Ingress, error) {
	return client.
		Extensions().
		Ingresses(client.Namespace).
		Get(name)
}

// Exist returns boolean value if ingress exists in k8s
func (client *Ingress) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates ingress in k8s
func (client *Ingress) Apply(resource *v1beta1.Ingress) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)

}

// Update updates existing ingress in k8s
func (client *Ingress) Update(resource *v1beta1.Ingress) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		Extensions().
		Ingresses(client.Namespace).
		Update(resource)
	return err
}

// Create creates new ingress in k8s
func (client *Ingress) Create(resource *v1beta1.Ingress) error {
	_, err := client.
		Extensions().
		Ingresses(client.Namespace).
		Create(resource)
	return err
}

// Destroy deletes ingress from the k8 cluster
func (client *Ingress) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	return client.Extensions().Ingresses(client.Namespace).Delete(name, options)
}

// List returns the list of k8s services maintained by pipeline
func (client *Ingress) List() ([]v1beta1.Ingress, error) {
	list, err := client.Extensions().Ingresses(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
