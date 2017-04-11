package reaper

import (
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestDeleteService(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: "sample",
			},
		},
		&v1beta1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      "abr",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"creator": "pipeline",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Env:          []v1.EnvVar{},
								VolumeMounts: []v1.VolumeMount{},
							},
						},
					},
				},
			},
		},
	)
	wrapper := &cluster.Cluster{
		Interface: c,
	}

	reaper := Reaper{
		Wrapper:   wrapper,
		Namespace: "sample",
	}

	cfg, _ := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	reaper.Cleanup(cfg)

	if d, err := wrapper.Extensions().Deployments("sample").Get("abr"); err == nil {
		t.Errorf("Expected deployment nil, got: %+v", d)
	}

	reaperFail := Reaper{
		Wrapper:   wrapper,
		Namespace: "nonexistent",
	}

	err := reaperFail.Cleanup(cfg)
	if err == nil {
		t.Errorf("Expected reaper cleanup to return error, got nil")
	}

}
