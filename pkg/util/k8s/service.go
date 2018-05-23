package k8s

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Service type actions on pvcs in k8s cluster
type Service struct {
	kubernetes.Interface
	Namespace string
}

// Get returns service object from the k8s by name
func (client *Service) Get(name string) (*v1.Service, error) {
	return client.Core().Services(client.Namespace).Get(name)
}

// Exist returns boolean value if pvc exists in k8s
func (client *Service) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates service in k8s
func (client *Service) Apply(resource *v1.Service) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new service in k8s
func (client *Service) Create(resource *v1.Service) error {
	_, err := client.
		Core().
		Services(client.Namespace).
		Create(resource)
	return err
}

// Update updates existing service in k8s
func (client *Service) Update(resource *v1.Service) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()
	resource.Spec.ClusterIP = current.Spec.ClusterIP

	_, err = client.
		Core().
		Services(client.Namespace).
		Update(resource)
	return err
}

// Destroy deletes service from the k8 cluster
func (client *Service) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	if client.deleteProtected(name) {
		return fmt.Errorf("Cannot destroy protected service %s", name)
		log.Errorf("Cannot destroy protected service %s", name)
	}
	return client.Core().Services(client.Namespace).Delete(name, options)
}

// List returns the list of k8s services maintained by pipeline
func (client *Service) List() ([]v1.Service, error) {
	list, err := client.Core().Services(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (client *Service) deleteProtected(name string) bool {
	svc, err := client.Get(name)
	if err == nil {
		meta := svc.GetObjectMeta()
		labels := meta.GetLabels()
		if len(labels) > 0 {
			if val, ok := labels["delete-protected"]; ok {
				if val == "yes" {
					return true
				}
			}
		}
	}
	return false
}
