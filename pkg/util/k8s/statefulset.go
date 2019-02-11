package k8s

import (
	"k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSet type actions on statefulset in k8s cluster
type StatefulSet struct {
	kubernetes.Interface
	Namespace string
}

// Get returns statefulset object from the k8s by name
func (client *StatefulSet) Get(name string) (*v1beta2.StatefulSet, error) {
	return client.AppsV1beta2().
		StatefulSets(client.Namespace).
		Get(name, getOptions())
}

// Exist returns boolean value if statefulset exists in k8s
func (client *StatefulSet) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates statefulset in k8s
func (client *StatefulSet) Apply(resource *v1beta2.StatefulSet) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)

}

// Update stateful set
func (client *StatefulSet) Update(resource *v1beta2.StatefulSet) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}

	current.Spec.Replicas = resource.Spec.Replicas
	_, err = client.
		AppsV1beta2().
		StatefulSets(client.Namespace).
		Update(current)

	/*resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		Apps().
		StatefulSets(client.Namespace).
		Update(resource)
	*/
	return err
}

// Create creates new statefulset in k8s
func (client *StatefulSet) Create(resource *v1beta2.StatefulSet) error {
	_, err := client.
		AppsV1beta2().
		StatefulSets(client.Namespace).
		Create(resource)
	return err
}

// Destroy deletes statefulset from the k8 cluster
func (client *StatefulSet) Destroy(name string) error {
	return client.Apps().StatefulSets(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *StatefulSet) List() ([]v1beta2.StatefulSet, error) {
	list, err := client.AppsV1beta2().StatefulSets(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
