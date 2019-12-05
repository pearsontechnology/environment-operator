package k8s

import (
	log "github.com/Sirupsen/logrus"
	extensions "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/client-go/rest"
)

// CustomResourceDefinition represents TPR API on the cluster
type CustomResourceDefinition struct {
	rest.Interface

	Namespace string
	Type      string
}

// Get retrieves PrsnExternalResource from the k8s using name
func (client *CustomResourceDefinition) Get(name string) (*extensions.PrsnExternalResource, error) {
	var rsc extensions.PrsnExternalResource

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
func (client *CustomResourceDefinition) Exist(name string) bool {
	rsc, _ := client.Get(name)
	return rsc != nil
}

// Apply creates or updates PrsnExternalResource in k8s
func (client *CustomResourceDefinition) Apply(resource *extensions.PrsnExternalResource) error {
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

// Create creates given tpr in
func (client *CustomResourceDefinition) Create(resource *extensions.PrsnExternalResource) error {
	if resource == nil {
		return nil
	}
	var result extensions.PrsnExternalResource
	return client.Interface.Post().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Update updates existing resource in k8s
func (client *CustomResourceDefinition) Update(resource *extensions.PrsnExternalResource) error {
	if resource == nil {
		return nil
	}
	var result extensions.PrsnExternalResource
	return client.Interface.Put().
		Resource(plural(client.Type)).
		Name(resource.ObjectMeta.Name).
		Namespace(client.Namespace).
		Body(resource).
		Do().Into(&result)
}

// Destroy deletes named resource
func (client *CustomResourceDefinition) Destroy(name string) error {
	var result extensions.PrsnExternalResource
	return client.Interface.Delete().
		Resource(plural(client.Type)).
		Namespace(client.Namespace).
		Name(name).Do().Into(&result)
}

// List returns a list of tprs. Depends on kind.
func (client *CustomResourceDefinition) List() ([]extensions.PrsnExternalResource, error) {
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
		if isVowel(rune(singular[len(singular)-2])) {
			plural = singular + "s"
			break
		}
		plural = singular[:len(singular)-1] + "ies"
	default:
		plural = singular + "s"
	}
	return plural
}

func isVowel(c rune) bool {
	vowels := []rune{'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U'}
	for _, value := range vowels {
		if value == c {
			return true
		}
	}
	return false
}
