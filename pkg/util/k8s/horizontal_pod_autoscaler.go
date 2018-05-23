package k8s

import (
	autoscale_v1 "k8s.io/api/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// HorizontalPodAutoscaler type actions in k8s cluster
type HorizontalPodAutoscaler struct {
	kubernetes.Interface
	Namespace string
}

// Get returns hpa object from k8s by name
func (client *HorizontalPodAutoscaler) Get(name string) (*autoscale_v1.HorizontalPodAutoscaler, error) {
	return client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).Get(name)
}

// Exist returns boolean value if hpa exists in k8s
func (client *HorizontalPodAutoscaler) Exist(name string) bool {
	_, err := client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).Get(name)
	return err == nil
}

// Apply updates or creates hpa in k8s
func (client *HorizontalPodAutoscaler) Apply(resource *autoscale_v1.HorizontalPodAutoscaler) error {
	if client.Exist(resource.Name) {
		return client.Update(resource)
	}
	return client.Create(resource)
}

// Create creates new hpa in k8s
func (client *HorizontalPodAutoscaler) Create(resource *autoscale_v1.HorizontalPodAutoscaler) error {
	var err error
	if *resource.Spec.MinReplicas != 0 {
		_, err = client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).Create(resource)
		return err
	}
	return err
}

// Update updates existing hpa in k8s
func (client *HorizontalPodAutoscaler) Update(resource *autoscale_v1.HorizontalPodAutoscaler) error {
	_, err := client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).Update(resource)
	return err
}

// Destroy deletes service from the k8 cluster
func (client *HorizontalPodAutoscaler) Destroy(name string) error {
	options := &v1.DeleteOptions{}
	return client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).Delete(name, options)
}

// List returns the list of k8s hpa
func (client *HorizontalPodAutoscaler) List() ([]autoscale_v1.HorizontalPodAutoscaler, error) {
	list, err := client.AutoscalingV1().HorizontalPodAutoscalers(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
