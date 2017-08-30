package k8s

import (
	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Secret is a client for interacting with secrets
type Secret struct {
	kubernetes.Interface
	Namespace string
}

// List returns the list of k8s secrets maintained by pipeline for provided client
func (client *Secret) List() ([]v1.Secret, error) {
	list, err := client.Core().Secrets(client.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// Check if the secret exists in the namespace
func (client *Secret) Exists(secretname string) bool {

	secrets, err := client.List()
	if err != nil {
		log.Error(err.Error())
	}
	found := false
	for _, sec := range secrets {
		if sec.ObjectMeta.Name == secretname {
			found = true
		}
	}
	return found
}
