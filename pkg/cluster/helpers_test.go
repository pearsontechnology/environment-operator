package cluster

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
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

func TestLivenessProbe(t *testing.T) {
	deployment := v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.IntOrString{
											IntVal: 8080,
										},
										HTTPHeaders: []v1.HTTPHeader{
											{
												Name:  "Custom-Header",
												Value: "Awesome",
											},
										},
									},
								},
								InitialDelaySeconds: 3,
								PeriodSeconds:       3,
							},
						},
					},
				},
			},
		},
	}

	r := livenessProbe(deployment)
	if r.HTTPGet.Path != "/healthz" {
		t.Errorf("Unexpected path in healthcehck. Expected /healthz, got: %s", r.HTTPGet.Path)
	}

	if r.HTTPGet.HTTPHeaders[0].Name != "Custom-Header" && r.HTTPGet.HTTPHeaders[0].Value != "Awesome" {
		t.Errorf("Unexpected header name %s or value %s", r.HTTPGet.HTTPHeaders[0].Name, r.HTTPGet.HTTPHeaders[0].Value)
	}

	if r.InitialDelaySeconds != 3 {
		t.Errorf("Unexpected initial delay: %d", r.InitialDelaySeconds)
	}

}

func TestReadinessProbe(t *testing.T) {
	deployment := v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											IntVal: 8080,
										},
									},
								},
								InitialDelaySeconds: 3,
								PeriodSeconds:       3,
								SuccessThreshold:    5,
								TimeoutSeconds:      6,
								FailureThreshold:    7,
							},
						},
					},
				},
			},
		},
	}

	r := readinessProbe(deployment)

	if r.TCPSocket.Port != 8080 {
		t.Errorf("Unexpected value for port %d", r.TCPSocket.Port)
	}

	if r.PeriodSeconds != 3 {
		t.Errorf("Unexpected value for period seconds %d", r.PeriodSeconds)
	}

	if r.InitialDelaySeconds != 3 {
		t.Errorf("Unexpected value for initial delay seconds %d", r.InitialDelaySeconds)
	}

	if r.SuccessThreshold != 5 {
		t.Errorf("Unexpected value for success threshold %d", r.SuccessThreshold)
	}

	if r.TimeoutSeconds != 6 {
		t.Errorf("Unexpected value for timeout seconds %d", r.TimeoutSeconds)
	}

	if r.FailureThreshold != 7 {
		t.Errorf("Unexpected value for failure threshold %d", r.FailureThreshold)
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

func TestReservedEnvVar(t *testing.T) {
	var tests = []struct {
		name     string
		expected bool
	}{
		{"POD_DEPLOYMENT_COLOUR", true},
		{"SOMEVAR", false},
	}

	for _, satest := range tests {
		e := v1.EnvVar{Name: satest.name, Value: ""}
		if isReservedEnvVar(e) != satest.expected {
			t.Errorf("unexpected result. expected %v got %v", satest.expected, isReservedEnvVar(e))
		}
	}
}
