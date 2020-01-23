package k8s

import (
	"os"

	log "github.com/Sirupsen/logrus"
	extensions "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/client-go/rest"
)

func init() {
	if os.Getenv("EXTERNAL_CRD_EXTERNAL_SECRETS_ENABLED") == "true" {
		ExternalSecretsEnabled = true
		log.Debugf("External CRD: external secrets enabled")
	}
}

var ExternalSecretsEnabled = false

// ExternalSecret represents ExternalSecret crd on the cluster
type ExternalSecret struct {
	rest.Interface

	Namespace string
	Type      string
}

// Get retrieves ExternalSecret from the k8s using name
func (client *ExternalSecret) Get(name string) (*extensions.ExternalSecret, error) {
	var rsc extensions.ExternalSecret

	err := client.Interface.Get().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Name(name).
		Do().Into(&rsc)

	if err != nil {
		log.Debugf("Got error on get: %s", err.Error())
		return nil, err
	}
	return &rsc, nil
}

// Exist checks if named resource exist in k8s cluster
func (client *ExternalSecret) Exist(name string) bool {
	rsc, _ := client.Get(name)
	return rsc != nil
}

// Apply creates or updates ExternalSecret in k8s
func (client *ExternalSecret) Apply(resource *extensions.ExternalSecret) error {
	if resource == nil {
		return nil
	}
	if client.Exist(resource.ObjectMeta.Name) {
		rsc, _ := client.Get(resource.ObjectMeta.Name)
		resource.ResourceVersion = rsc.GetResourceVersion()
		log.Debugf("Updating CRD resource: %s", resource.ObjectMeta.Name)
		ret := client.Update(resource)
		if ret != nil {
			log.Debugf("CRD: Got error on update: %s", ret.Error())
		}
		return ret
	}
	log.Debugf("Creating CRD resource: %s", resource.ObjectMeta.Name)
	ret := client.Create(resource)
	if ret != nil {
		log.Debugf("TPR: Got error on create: %s", ret.Error())
	}
	return ret
}

// Create creates given ExternalSecret in k8s
func (client *ExternalSecret) Create(resource *extensions.ExternalSecret) error {
	if resource == nil {
		return nil
	}
	var result extensions.ExternalSecret
	return client.Interface.Post().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Update updates existing resource in k8s
func (client *ExternalSecret) Update(resource *extensions.ExternalSecret) error {
	if resource == nil {
		return nil
	}
	var result extensions.ExternalSecret
	return client.Interface.Put().
		Resource(plural(client.Type)).
		Name(resource.ObjectMeta.Name).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Destroy deletes named resource
func (client *ExternalSecret) Destroy(name string) error {
	var result extensions.ExternalSecret
	return client.Interface.Delete().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Name(name).Do().Into(&result)
}

// List returns a list of ExternalSecret.
func (client *ExternalSecret) List() ([]extensions.ExternalSecret, error) {
	var result extensions.ExternalSecretList
	err := client.Interface.Get().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Do().Into(&result)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
