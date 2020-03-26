package k8s

import (
	log "github.com/Sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PersistentVolumeClaim type actions on pvcs in k8s cluster
type PersistentVolumeClaim struct {
	kubernetes.Interface
	Namespace string
}

// Get returns pvc object from the k8s by name
func (client *PersistentVolumeClaim) Get(name string) (*v1.PersistentVolumeClaim, error) {
	return client.CoreV1().PersistentVolumeClaims(client.Namespace).Get(name, metav1.GetOptions{})
}

// Exist returns boolean value if pvc exists in k8s
func (client *PersistentVolumeClaim) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates pvc in k8s
func (client *PersistentVolumeClaim) Apply(resource *v1.PersistentVolumeClaim) error {
	if resource == nil {
		return nil
	}
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new ingress in k8s
func (client *PersistentVolumeClaim) Create(resource *v1.PersistentVolumeClaim) error {
	if resource == nil {
		return nil
	}
	_, err := client.
		CoreV1().
		PersistentVolumeClaims(client.Namespace).
		Create(resource)
	return err
}

// Update updates existing ingress in k8s
func (client *PersistentVolumeClaim) Update(resource *v1.PersistentVolumeClaim) error {
	if resource == nil {
		return nil
	}
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()
	resource.Spec.VolumeName = current.Spec.VolumeName

	log.Warningf("attemting to update volume \"%s\", service \"%s\", but PVC Spec is immutable so this may fail.", current.ObjectMeta.Name, current.ObjectMeta.Labels["deployment"])

	_, err = client.
		CoreV1().
		PersistentVolumeClaims(client.Namespace).
		Update(resource)

	if err == nil {
		log.Warningf("succesfully  updated volume \"%s\", service \"%s\".", current.ObjectMeta.Name, current.ObjectMeta.Labels["deployment"])
	}

	return err
}

// Destroy deletes pvc from the k8 cluster
func (client *PersistentVolumeClaim) Destroy(name string) error {
	return client.CoreV1().PersistentVolumeClaims(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *PersistentVolumeClaim) List() ([]v1.PersistentVolumeClaim, error) {
	list, err := client.CoreV1().PersistentVolumeClaims(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
