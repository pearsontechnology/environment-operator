package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Namespace is a client for interacting with namespaces
type Namespace struct {
	Interface kubernetes.Interface
	Namespace string
}

// Get returns namespace object from the k8s by name
func (client *Namespace) Get() (*v1.Namespace, error) {
	return client.Interface.Core().Namespaces().Get(client.Namespace)
}
