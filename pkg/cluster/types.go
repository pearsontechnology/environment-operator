package cluster

import "k8s.io/client-go/kubernetes"

// Cluster wraps low level kubernetes api requests to an object easier
// to interact with
type Cluster struct {
	kubernetes.Interface
	TestMode bool
}
