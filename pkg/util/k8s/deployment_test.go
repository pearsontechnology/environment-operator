package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestDeploymentGet(t *testing.T) {
	d := createDeployment()
	if _, err := d.Get("test"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m, err := d.Get("nonexistent"); err == nil {
		t.Errorf("Unexpected deployment: %v", m)
	}
}

func TestDeploymentExist(t *testing.T) {
	d := createDeployment()
	var saTests = []struct {
		DeploymentName string
		Expected       bool
		Message        string
	}{
		{"test", true, "Existing deployment not found"},
		{"nonexistent", false, "Unexpected deployment 'nonexistent'"},
	}

	for _, sTest := range saTests {
		if d.Exist(sTest.DeploymentName) != sTest.Expected {
			t.Error(sTest.Message)
		}
	}
}

func TestDeploymentApplyNew(t *testing.T) {
	d := createDeployment()
	newDeployment := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "new",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
				"version": "0.0.1",
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "test:0.0.1",
						},
					},
				},
			},
		},
	}
	if err := d.Apply(newDeployment); err != nil {
		t.Errorf("Unexpected error applying deployment: %s", err.Error())
	}
	m, err := d.Get("new")
	if err != nil {
		t.Errorf("Applied deployment not found")
	}
	if m.ObjectMeta.Labels["version"] != "0.0.1" {
		t.Errorf("Invalid version label. Expected 0.0.1, got %s", m.ObjectMeta.Labels["version"])
	}
}

func TestDeploymentApplyExisting(t *testing.T) {
	d := createDeployment()
	existingDeployment := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "sample",
			Labels: map[string]string{
				"creator": "pipeline",
				"version": "0.2",
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "asd",
						},
					},
				},
			},
		},
	}
	if err := d.Apply(existingDeployment); err != nil {
		t.Errorf("Unexpected error applying deployment: %s", err.Error())
	}

	m, _ := d.Get("test")
	if m.ObjectMeta.Labels["version"] != "0.2" {
		t.Errorf("Update during apply failed, version not applied: %s", m.ObjectMeta.Labels["version"])
	}
	if m.Spec.Template.Spec.Containers[0].Image != "asd" {
		t.Errorf("Invalid image name. Expected asd, got: %s", m.Spec.Template.Spec.Containers[0].Image)
	}
}

func createDeployment() Deployment {
	return Deployment{
		Interface: createSimpleDeploymentClient(),
		Namespace: "sample",
	}
}

func createSimpleDeploymentClient() *fake.Clientset {
	replicaCount := int32(1)
	return fake.NewSimpleClientset(
		&v1beta1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1beta1.DeploymentSpec{
				Replicas: &replicaCount,
				Template: v1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"creator": "pipeline",
							"version": "1",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "someimage",
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "test",
										MountPath: "/tmp/blah",
										ReadOnly:  true,
									},
								},
							},
						},
						Volumes: []v1.Volume{
							{
								Name: "test",
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
										ClaimName: "test",
									},
								},
							},
						},
					},
				},
			},
		},
	)
}
