package reaper

import (
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	fakecrd "github.com/pearsontechnology/environment-operator/pkg/util/k8s/fake"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeleteService(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "sample",
			},
		},
		&v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "abr",
				Namespace: "sample",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
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

	crdcli := fakecrd.CRDClient("prsn.io", "v1")

	wrapper := &cluster.Cluster{
		Interface: c,
		CRDClient: crdcli,
	}

	reaper := Reaper{
		Wrapper:   wrapper,
		Namespace: "sample",
	}

	cfg, _ := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	reaper.Cleanup(cfg)

	if d, err := wrapper.Extensions().Deployments("sample").Get("abr", metav1.GetOptions{}); err == nil {
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
