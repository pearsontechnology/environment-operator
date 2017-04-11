package k8s

import (
	log "github.com/Sirupsen/logrus"
	extensions "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/client-go/rest"
)

// ThirdPartyResource represents TPR API on the cluster
type ThirdPartyResource struct {
	rest.Interface

	Namespace string
	Type      string
}

// Get retrieves PrsnExternalResource from the k8s using name
func (client *ThirdPartyResource) Get(name string) (*extensions.PrsnExternalResource, error) {
	var rsc extensions.PrsnExternalResource

	err := client.Interface.Get().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Name(name).
		Do().Into(&rsc)

	if err != nil {
		log.Debugf("Got error on get: %s", err.Error())
	}
	return &rsc, err
}

// Exist checks if named resource exist in k8s cluster
func (client *ThirdPartyResource) Exist(name string) bool {
	rsc, _ := client.Get(name)
	return rsc != nil
}

// Apply creates or updates PrsnExternalResource in k8s
func (client *ThirdPartyResource) Apply(resource *extensions.PrsnExternalResource) error {
	if client.Exist(resource.ObjectMeta.Name) {
		log.Debugf("Updating tpr resource: %s", resource.ObjectMeta.Name)
		ret := client.Update(resource)
		if ret != nil {
			log.Debugf("TPR: Got error on update: %s", ret.Error())
		}
		return ret
	}
	log.Debugf("Creating tpr resource: %s", resource.ObjectMeta.Name)
	ret := client.Create(resource)
	if ret != nil {
		log.Debugf("TPR: Got error on create: %s", ret.Error())
	}
	return ret
}

// Create creates given tpr in
func (client *ThirdPartyResource) Create(resource *extensions.PrsnExternalResource) error {
	var result extensions.PrsnExternalResource
	return client.Interface.Post().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Update updates existing resource in k8s
func (client *ThirdPartyResource) Update(resource *extensions.PrsnExternalResource) error {
	var result extensions.PrsnExternalResource
	return client.Put().
		Resource(plural(client.Type)).
		Name(resource.ObjectMeta.Name).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Destroy deletes named resource
func (client *ThirdPartyResource) Destroy(name string) error {
	var result extensions.PrsnExternalResource
	return client.Interface.Delete().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Name(name).Do().Into(&result)
}

// List returns a list of tprs. Depends on kind.
func (client *ThirdPartyResource) List() ([]extensions.PrsnExternalResource, error) {
	var result extensions.PrsnExternalResourceList
	err := client.Interface.Get().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Do().Into(&result)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func plural(singular string) string {
	var plural string

	switch string(singular[len(singular)-1]) {
	case "s", "x":
		plural = singular + "es"
	case "y":
		plural = singular[:len(singular)-1] + "ies"
	default:
		plural = singular + "s"
	}
	return plural
}
