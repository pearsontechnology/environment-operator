package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Namespace is a client for interacting with namespaces
type Namespace struct {
	Interface kubernetes.Interface
	Namespace string
}

// Get returns namespace object from the k8s by name
func (client *Namespace) Get() (*v1.Namespace, error) {
	return client.Interface.CoreV1().Namespaces().Get(client.Namespace, metav1.GetOptions{})
}
