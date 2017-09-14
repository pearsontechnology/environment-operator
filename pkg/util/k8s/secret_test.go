package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

func TestSecretExists(t *testing.T) {
	client := createSecret()

	var secretTests = []struct {
		Name     string
		Expected bool
		Message  string
	}{
		{"test-secret", true, "Existing secret not found"},
		{"nonexistent", false, "Unexpected secret 'nonexistent'"},
	}

	for _, sTest := range secretTests {
		if client.Exists(sTest.Name) != sTest.Expected {
			t.Error(sTest.Message)
		}
	}
}

//Test that secrets can be pulled from a secrets client using the List function
func TestSecretList(t *testing.T) {
	client := createSecret()

	s, err := client.List()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if len(s) != 1 {
		t.Errorf("Unexpected count of secrets, expected: 1, got: %d", len(s))
	}
}

func createSecret() Secret {
	return Secret{
		Interface: createSimpleSecretClient(),
		Namespace: "sample",
	}
}

func createSimpleSecretClient() *fake.Clientset {
	return fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-secret",
				Namespace: "sample",
			},
		},
	)
}
