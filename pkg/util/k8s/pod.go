package k8s

import (
	"bytes"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Pod type actions on pods in k8s cluster
type Pod struct {
	kubernetes.Interface
	Namespace string
}

// GetLogs returns pod's logs as a string
func (client *Pod) GetLogs(name string) (string, error) {

	reader, err := client.CoreV1().Pods(client.Namespace).GetLogs(name, logOptions()).Stream()
	if err != nil {
		return "", err
	}
	defer reader.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String(), err
}

// List returns the list of k8s services maintained by pipeline
func (client *Pod) List() ([]v1.Pod, error) {
	list, err := client.CoreV1().Pods(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
