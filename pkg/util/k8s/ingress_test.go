package k8s

import (
	"testing"

	netwk_v1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressGet(t *testing.T) {
	client := createIngress()
	if _, err := client.Get("test"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m, err := client.Get("nonexistent"); err == nil {
		t.Errorf("Unexpected ingress: %v", m)
	}
}

func TestIngressExist(t *testing.T) {
	client := createIngress()
	var saTests = []struct {
		IngressName string
		Expected    bool
		Message     string
	}{
		{"test", true, "Existing ingress not found"},
		{"nonexistent", false, "Unexpected ingress 'nonexistent'"},
	}

	for _, sTest := range saTests {
		if client.Exist(sTest.IngressName) != sTest.Expected {
			t.Error(sTest.Message)
		}
	}
}

func TestIngressApplyNew(t *testing.T) {
	client := createIngress()
	newResource := &netwk_v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Apply(newResource); err != nil {
		t.Errorf("Unexpected error applying ingress: %s", err.Error())
	}
	_, err := client.Get("new")
	if err != nil {
		t.Errorf("Applied ingress not found")
	}
}

func TestIngressApplyExisting(t *testing.T) {
	client := createIngress()
	existing := &netwk_v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
			},
		},
	}
	if err := client.Apply(existing); err != nil {
		t.Errorf("Unexpected error applying ingress: %s", err.Error())
	}

	// m, _ := client.Get("test")
	// if m.ObjectMeta.Labels["version"] != "0.2" {
	// 	t.Errorf("Update during apply failed, version not applied: %s", m.ObjectMeta.Labels["version"])
	// }
}

func createIngress() Ingress {
	return Ingress{
		Interface: createSimpleIngressClient(),
		Namespace: "sample",
	}
}

func createSimpleIngressClient() *fake.Clientset {
	return fake.NewSimpleClientset(
		&netwk_v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
		},
	)
}
