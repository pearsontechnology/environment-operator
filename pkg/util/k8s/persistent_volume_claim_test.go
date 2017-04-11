package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

func TestPVCGet(t *testing.T) {
	client := createPVC()
	if _, err := client.Get("test"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m, err := client.Get("nonexistent"); err == nil {
		t.Errorf("Unexpected pvc: %v", m)
	}
}

func TestPVCExist(t *testing.T) {
	client := createPVC()
	var saTests = []struct {
		IngressName string
		Expected    bool
		Message     string
	}{
		{"test", true, "Existing pvc not found"},
		{"nonexistent", false, "Unexpected pvc 'nonexistent'"},
	}

	for _, sTest := range saTests {
		if client.Exist(sTest.IngressName) != sTest.Expected {
			t.Error(sTest.Message)
		}
	}
}

func TestPVCApplyNew(t *testing.T) {
	client := createPVC()
	newResource := &v1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "new",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Apply(newResource); err != nil {
		t.Errorf("Unexpected error applying pvc: %s", err.Error())
	}
	_, err := client.Get("new")
	if err != nil {
		t.Errorf("Applied pvc not found")
	}
}

func TestPVCApplyExisting(t *testing.T) {
	client := createPVC()
	existing := &v1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
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

func TestPVCUpdateNonexisting(t *testing.T) {
	client := createPVC()
	resource := &v1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
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

func createPVC() PersistentVolumeClaim {
	return PersistentVolumeClaim{
		Interface: createPVCClient(),
		Namespace: "sample",
	}
}

func createPVCClient() *fake.Clientset {
	return fake.NewSimpleClientset(
		&v1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				VolumeName: "test",
			},
		},
	)
}
