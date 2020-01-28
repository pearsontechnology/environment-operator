package k8s

import (
	autoscale_v2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HorizontalPodAutoscaler type actions in k8s cluster
type HorizontalPodAutoscaler struct {
	kubernetes.Interface
	Namespace string
}

// Get returns hpa object from k8s by name
func (client *HorizontalPodAutoscaler) Get(name string) (*autoscale_v2beta2.HorizontalPodAutoscaler, error) {
	return client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).Get(name, metav1.GetOptions{})
}

// Exist returns boolean value if hpa exists in k8s
func (client *HorizontalPodAutoscaler) Exist(name string) bool {
	_, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).Get(name, metav1.GetOptions{})
	return err == nil
}

// Apply updates or creates hpa in k8s
func (client *HorizontalPodAutoscaler) Apply(resource *autoscale_v2beta2.HorizontalPodAutoscaler) error {
	if resource == nil {
		return nil
	}
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new hpa in k8s
func (client *HorizontalPodAutoscaler) Create(resource *autoscale_v2beta2.HorizontalPodAutoscaler) error {
	var err error
	if resource == nil {
		return nil
	}

	if *resource.Spec.MinReplicas != 0 {
		_, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).Create(resource)
		return err
	}
	return err
}

// Update updates existing hpa in k8s
func (client *HorizontalPodAutoscaler) Update(resource *autoscale_v2beta2.HorizontalPodAutoscaler) error {
	if resource == nil {
		return nil
	}
	_, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).Update(resource)
	return err
}

// Destroy deletes service from the k8 cluster
func (client *HorizontalPodAutoscaler) Destroy(name string) error {
	return client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).Delete(name, &metav1.DeleteOptions{})
}

// List returns the list of k8s hpa
func (client *HorizontalPodAutoscaler) List() ([]autoscale_v2beta2.HorizontalPodAutoscaler, error) {
	list, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
