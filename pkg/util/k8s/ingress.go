package k8s

import (
	netwk_v1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Ingress type actions on ingresses in k8s cluster
type Ingress struct {
	kubernetes.Interface
	Namespace string
}

// Get returns ingress object from the k8s by name
func (client *Ingress) Get(name string) (*netwk_v1beta1.Ingress, error) {
	return client.
		NetworkingV1beta1().
		Ingresses(client.Namespace).
		Get(name, getOptions())
}

// Exist returns boolean value if ingress exists in k8s
func (client *Ingress) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates ingress in k8s
func (client *Ingress) Apply(resource *netwk_v1beta1.Ingress) error {
	if resource == nil {
		return nil
	}
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)

}

// Update updates existing ingress in k8s
func (client *Ingress) Update(resource *netwk_v1beta1.Ingress) error {
	if resource == nil {
		return nil
	}
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		NetworkingV1beta1().
		Ingresses(client.Namespace).
		Update(resource)
	return err
}

// Create creates new ingress in k8s
func (client *Ingress) Create(resource *netwk_v1beta1.Ingress) error {
	if resource == nil {
		return nil
	}
	_, err := client.
		NetworkingV1beta1().
		Ingresses(client.Namespace).
		Create(resource)
	return err
}

// Destroy deletes ingress from the k8 cluster
func (client *Ingress) Destroy(name string) error {
	return client.NetworkingV1beta1().Ingresses(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *Ingress) List() ([]netwk_v1beta1.Ingress, error) {
	list, err := client.NetworkingV1beta1().Ingresses(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
