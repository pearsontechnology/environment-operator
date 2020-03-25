package k8s

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestServiceGet(t *testing.T) {
	client := createService()
	if _, err := client.Get("test"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m, err := client.Get("nonexistent"); err == nil {
		t.Errorf("Unexpected pvc: %v", m)
	}
}

func TestServiceExist(t *testing.T) {
	client := createService()
	var saTests = []struct {
		Name     string
		Expected bool
		Message  string
	}{
		{"test", true, "Existing service not found"},
		{"nonexistent", false, "Unexpected service 'nonexistent'"},
	}

	for _, sTest := range saTests {
		if client.Exist(sTest.Name) != sTest.Expected {
			t.Error(sTest.Message)
		}
	}
}

func TestServiceApplyNew(t *testing.T) {
	client := createService()
	newResource := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Apply(newResource); err != nil {
		t.Errorf("Unexpected error applying service: %s", err.Error())
	}
	_, err := client.Get("new")
	if err != nil {
		t.Errorf("Applied service not found")
	}
}

func TestServiceApplyExisting(t *testing.T) {
	client := createService()
	existing := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Apply(existing); err != nil {
		t.Errorf("Unexpected error applying pvc: %s", err.Error())
	}
}

func TestServiceUpdateNonexisting(t *testing.T) {
	client := createService()
	resource := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nonexisting",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Update(resource); err == nil {
		t.Error("Error should be raised, but got nil")
	}
}

func TestServiceDestroyExisting(t *testing.T) {
	client := createService()
	err := client.Destroy("test")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
}

func TestServiceDestroyNonExisting(t *testing.T) {
	client := createService()
	err := client.Destroy("nonexisting")
	if err == nil {
		t.Errorf("Unexpected error nil")
	}
}

func TestServiceList(t *testing.T) {
	client := createService()
	s, err := client.List()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if len(s) != 1 {
		t.Errorf("Unexpected count of services, expected: 1, got: %d", len(s))
	}
}

func TestServiceListWrongNs(t *testing.T) {
	client := createService()
	client.Namespace = "other"
	c, _ := client.List()
	if len(c) != 0 {
		t.Error(c)
	}
}

func createService() Service {
	f := fake.NewSimpleClientset(
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
		},
	)
	return Service{
		Interface: f,
		Namespace: "sample",
	}
}
