package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Client is a top level struct, wrapping all other clients
type Client struct {
	Interface kubernetes.Interface
	Namespace string
	CRDClient rest.Interface
}

// ClientForNamespace configures REST client to operate in a given namespace
func ClientForNamespace(ns string) (*Client, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	restcli, err := CRDClient(nil)
	if err != nil {
		return nil, err
	}

	return &Client{Interface: clientset, Namespace: ns, CRDClient: restcli}, nil
}

// CRDClient returns rest.RESTClient for CustomResourceDefinitions
func CRDClient(groupVersion *schema.GroupVersion) (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	config.GroupVersion = &schema.GroupVersion{
		Group:   "prsn.io",
		Version: "v1",
	}

	if groupVersion != nil {
		config.GroupVersion = groupVersion
	}

	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	return rest.RESTClientFor(config)
}

// Service builds Service client
func (c *Client) Service() *Service {
	return &Service{Interface: c.Interface, Namespace: c.Namespace}
}

// Deployment builds Deployment client
func (c *Client) Deployment() *Deployment {
	return &Deployment{Interface: c.Interface, Namespace: c.Namespace}
}

// HorizontalPodAutoscaler builds HPA client
func (c *Client) HorizontalPodAutoscaler() *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{Interface: c.Interface, Namespace: c.Namespace}
}

// ConfigMap builds ConfigMap client
func (c *Client) ConfigMap() *ConfigMap {
	return &ConfigMap{Interface: c.Interface, Namespace: c.Namespace}
}

// Secret builds Secrets client
func (c *Client) Secret() *Secret {
	return &Secret{Interface: c.Interface, Namespace: c.Namespace}
}

// PVC builds PersistentVolumeClaim client
func (c *Client) PVC() *PersistentVolumeClaim {
	return &PersistentVolumeClaim{Interface: c.Interface, Namespace: c.Namespace}
}

// Pod builds Pod client
func (c *Client) Pod() *Pod {
	return &Pod{Interface: c.Interface, Namespace: c.Namespace}
}

// Ingress builds Ingress client
func (c *Client) Ingress() *Ingress {
	return &Ingress{Interface: c.Interface, Namespace: c.Namespace}
}

// StatefulSet builds Statefulset client
func (c *Client) StatefulSet() *StatefulSet {
	return &StatefulSet{Interface: c.Interface, Namespace: c.Namespace}
}

// Ns builds Ingress client
func (c *Client) Ns() *Namespace {
	return &Namespace{Interface: c.Interface, Namespace: c.Namespace}
}

// ConfigMap builds ConfigMap client
func (c *Client) Job() *Job {
	return &Job{Interface: c.Interface, Namespace: c.Namespace}
}

// CustomResourceDefinition builds CRD client
func (c *Client) CustomResourceDefinition(kind string) *CustomResourceDefinition {

	return &CustomResourceDefinition{
		Interface: c.CRDClient,
		Namespace: c.Namespace,
		Type:      kind,
	}
}

// ExternalSecret builds ExternalSec
func (c *Client) ExternalSecret() *ExternalSecret {

	return &ExternalSecret{
		Interface: c.CRDClient,
		Namespace: c.Namespace,
		Type:      "ExternalSecret",
	}
}

func listOptions() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "creator=pipeline",
	}
}

func getOptions() metav1.GetOptions {
	return metav1.GetOptions{}
}

func logOptions() *v1.PodLogOptions {
	return &v1.PodLogOptions{
		//SinceSeconds: &[]int64{300}[0], //Gets last 5 minutes of logs
		TailLines:  &[]int64{500}[0], //Retrieve last 500 lines from pod log
		Timestamps: true,             //Add timestamp to each line in the log
	}
}
