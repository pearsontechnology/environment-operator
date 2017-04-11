package k8s

import "k8s.io/client-go/kubernetes"
import "k8s.io/client-go/pkg/api/v1"

// PersistentVolumeClaim type actions on pvcs in k8s cluster
type PersistentVolumeClaim struct {
	kubernetes.Interface
	Namespace string
}

// Get returns pvc object from the k8s by name
func (client *PersistentVolumeClaim) Get(name string) (*v1.PersistentVolumeClaim, error) {
	return client.Core().PersistentVolumeClaims(client.Namespace).Get(name)
}

// Exist returns boolean value if pvc exists in k8s
func (client *PersistentVolumeClaim) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates pvc in k8s
func (client *PersistentVolumeClaim) Apply(resource *v1.PersistentVolumeClaim) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new ingress in k8s
func (client *PersistentVolumeClaim) Create(resource *v1.PersistentVolumeClaim) error {
	_, err := client.
		Core().
		PersistentVolumeClaims(client.Namespace).
		Create(resource)
	return err
}

// Update updates existing ingress in k8s
func (client *PersistentVolumeClaim) Update(resource *v1.PersistentVolumeClaim) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		Core().
		PersistentVolumeClaims(client.Namespace).
		Update(resource)
	return err
}

// Destroy deletes pvc from the k8 cluster
func (client *PersistentVolumeClaim) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	return client.Core().PersistentVolumeClaims(client.Namespace).Delete(name, options)
}

// List returns the list of k8s services maintained by pipeline
func (client *PersistentVolumeClaim) List() ([]v1.PersistentVolumeClaim, error) {
	list, err := client.Core().PersistentVolumeClaims(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
