package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	appsv1 "k8s.io/api/apps/v1"
)

// StatefulSet type actions on statefulset in k8s cluster
type StatefulSet struct {
	kubernetes.Interface
	Namespace string
}

// Get returns statefulset object from the k8s by name
func (client *StatefulSet) Get(name string) (*appsv1.StatefulSet, error) {
	return client.Apps().
		StatefulSets(client.Namespace).
		Get(name)
}

// Exist returns boolean value if statefulset exists in k8s
func (client *StatefulSet) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates statefulset in k8s
func (client *StatefulSet) Apply(resource *v1beta1.StatefulSet) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)

}

//Only allow updates to the replica field of a statefulset
// K8s 1.5.7 does not support upgrades to Statefulsets. Forbidden: updates to statefulset spec for fields other than 'replicas' are forbidden.
//Other updates will be ignored until the following are implemented and we upgrade our k8s
//https://github.com/kubernetes/kubernetes/issues/41015
//https://github.com/kubernetes/kubernetes/pull/46669

func (client *StatefulSet) Update(resource *v1beta1.StatefulSet) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}

	current.Spec.Replicas = resource.Spec.Replicas

	_, err = client.
		Apps().
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
func (client *StatefulSet) Create(resource *v1beta1.StatefulSet) error {
	_, err := client.
		Apps().
		StatefulSets(client.Namespace).
		Create(resource)
	return err
}

// Destroy deletes statefulset from the k8 cluster
func (client *StatefulSet) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	return client.Apps().StatefulSets(client.Namespace).Delete(name, options)
}

// List returns the list of k8s services maintained by pipeline
func (client *StatefulSet) List() ([]v1beta1.StatefulSet, error) {
	list, err := client.Apps().StatefulSets(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
