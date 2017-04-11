package cluster

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestHealthCheck(t *testing.T) {
	deployment := v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"ls"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	r := healthCheck(deployment)
	if r.Command[0] != "ls" {
		t.Errorf("Unexpected command in healthcehck. Expected ls, got: %s", r.Command[0])
	}

	if r.InitialDelay != 0 {
		t.Errorf("Unexpected initial delay: %d", r.InitialDelay)
	}

}

func TestGetAccessModesAsString(t *testing.T) {
	modes := []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce, v1.ReadOnlyMany, v1.ReadWriteMany,
	}
	str := getAccessModesAsString(modes)
	if str != "ReadWriteOnce,ReadOnlyMany,ReadWriteMany" {
		t.Errorf("Wrong mode: %s", str)
	}
}
