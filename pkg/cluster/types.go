package cluster

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Cluster wraps low level kubernetes api requests to an object easier
// to interact with
type Cluster struct {
	kubernetes.Interface
	TPRClient rest.Interface
}
