package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

/*
Currently disabled. Not able to test Log Streams due to SegFault Issue: https://github.com/kubernetes/client-go/issues/196
func TestPodGetLogs(t *testing.T) {
	client := createPod()
	if _, err := client.GetLogs("test"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m, err := client.GetLogs("nonexistent"); err == nil {
		t.Errorf("Unexpected Pod: %v", m)
	}

}*/

func TestPodList(t *testing.T) {
	client := createPod()
	s, err := client.List()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if len(s) != 1 {
		t.Errorf("Unexpected count of Pod Services, expected: 1, got: %d", len(s))
	}
}

func createPod() Pod {
	f := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
		},
	)
	return Pod{
		Interface: f,
		Namespace: "sample",
	}
}
