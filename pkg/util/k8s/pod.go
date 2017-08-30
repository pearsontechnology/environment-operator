package k8s

import "k8s.io/client-go/kubernetes"
import "k8s.io/client-go/pkg/api/v1"
import (
	"bytes"
)

// Pod type actions on pods in k8s cluster
type Pod struct {
	kubernetes.Interface
	Namespace string
}

func (client *Pod) GetLogs(name string) (string, error) {

	reader, err := client.Core().Pods(client.Namespace).GetLogs(name, logOptions()).Stream()
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
	list, err := client.Core().Pods(client.Namespace).List(listOptions())
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
