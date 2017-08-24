package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Deployment type actions on ingresses in k8s cluster
type Deployment struct {
	kubernetes.Interface
	Namespace string
}

// Get returns deployment object from the k8s by name
func (client *Deployment) Get(name string) (*v1beta1.Deployment, error) {
	return client.
		Extensions().
		Deployments(client.Namespace).
		Get(name)
}

// Exist returns boolean value if deployment exists in k8s
func (client *Deployment) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates deployment in k8s
func (client *Deployment) Apply(deployment *v1beta1.Deployment) error {
	if client.Exist(deployment.Name) {
		return client.Update(deployment)
	}
	return client.Create(deployment)
}

// Update updates existing deployment in k8s
func (client *Deployment) Update(deployment *v1beta1.Deployment) error {
	current, err := client.Get(deployment.Name)
	if err != nil {
		return err
	}
	deployment.ResourceVersion = current.GetResourceVersion()
	if deployment.ObjectMeta.Labels["version"] == "" {
		deployment.ObjectMeta.Labels["version"] = current.ObjectMeta.Labels["version"]
	}

	if len(current.Spec.Template.Spec.Containers) > 0 &&
		len(deployment.Spec.Template.Spec.Containers) > 0 &&
		deployment.Spec.Template.Spec.Containers[0].Image == "" {
		deployment.Spec.Template.Spec.Containers[0].Image = current.Spec.Template.Spec.Containers[0].Image
	}
	_, err = client.
		Extensions().
		Deployments(client.Namespace).
		Update(deployment)
	return err
}

// Create creates new deployment in k8s
func (client *Deployment) Create(deployment *v1beta1.Deployment) error {
	var err error
	if len(deployment.Spec.Template.Spec.Containers) > 0 &&
		deployment.Spec.Template.Spec.Containers[0].Image != "" {
		_, err = client.
			Extensions().
			Deployments(client.Namespace).
			Create(deployment)
		return err
	}
	return fmt.Errorf("Error creating deployment %s; image not set", deployment.Name)
}

// Destroy deletes deployment from the k8 cluster
func (client *Deployment) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	return client.Extensions().Deployments(client.Namespace).Delete(name, options)
}

// List returns the list of k8s services maintained by pipeline
func (client *Deployment) List() ([]v1beta1.Deployment, error) {
	list, err := client.Extensions().Deployments(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
