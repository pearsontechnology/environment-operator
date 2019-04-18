package k8s

import (
	v1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CronJob type actions on cronjobs in k8s cluster
type CronJob struct {
	kubernetes.Interface
	Namespace string
}

// Get returns service object from the k8s by name
func (client *CronJob) Get(name string) (*v1beta1.CronJob, error) {
	return client.BatchV1beta1().CronJobs(client.Namespace).Get(name, getOptions())
}

// Exist returns boolean value if pvc exists in k8s
func (client *CronJob) Exist(name string) bool {
	_, err := client.Get(name)
	return err == nil
}

// Apply updates or creates service in k8s
func (client *CronJob) Apply(resource *v1beta1.CronJob) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new service in k8s
func (client *CronJob) Create(resource *v1beta1.CronJob) error {
	_, err := client.
		BatchV1beta1().
		CronJobs(client.Namespace).
		Create(resource)
	return err
}

// Update updates existing service in k8s
func (client *CronJob) Update(resource *v1beta1.CronJob) error {
	current, err := client.Get(resource.Name)
	if err != nil {
		return err
	}
	resource.ResourceVersion = current.GetResourceVersion()

	_, err = client.
		BatchV1beta1().
		CronJobs(client.Namespace).
		Update(resource)
	return err
}

// Destroy deletes service from the k8 cluster
func (client *CronJob) Destroy(name string) error {
	return client.BatchV1beta1().CronJobs(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s services maintained by pipeline
func (client *CronJob) List() ([]v1beta1.CronJob, error) {
	list, err := client.BatchV1beta1().CronJobs(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
