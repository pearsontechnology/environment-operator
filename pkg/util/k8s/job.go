package k8s

import (
	v1batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Job type actions on pvcs in k8s cluster
type Job struct {
	kubernetes.Interface
	Namespace string
}

// Get returns service object from the k8s by name
func (client *Job) Get(name string) (*v1batch.Job, error) {
	return client.BatchV1().Jobs(client.Namespace).Get(name, getOptions())
}

// Exist returns boolean value if pvc exists in k8s
func (client *Job) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates service in k8s
func (client *Job) Apply(resource *v1batch.Job) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new service in k8s
func (client *Job) Create(resource *v1batch.Job) error {
	_, err := client.
		BatchV1().
		Jobs(client.Namespace).
		Create(resource)
	return err
}

// Update updates existing job in k8s
func (client *Job) Update(job *v1batch.Job) error {
	current, err := client.Get(job.Name)
	if err != nil {
		return err
	}
	job.ResourceVersion = current.GetResourceVersion()
	_, err = client.
		BatchV1().
		Jobs(client.Namespace).
		Update(job)
	return err
}

// Destroy deletes service from the k8 cluster
func (client *Job) Destroy(name string) error {
	return client.
		BatchV1().
		Jobs(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *Job) List() ([]v1batch.Job, error) {
	list, err := client.
		BatchV1().
		Jobs(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
