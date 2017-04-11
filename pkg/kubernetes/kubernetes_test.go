package kubernetes

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func testFullBitesizeEnvironment(t *testing.T) {
	capacity, _ := resource.ParseQuantity("59G")
	validLabels := map[string]string{"creator": "pipeline"}
	nsLabels := map[string]string{"environment": "Development"}
	replicaCount := int32(1)

	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name:   "test",
				Labels: nsLabels,
			},
		},
		&v1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ts",
				Namespace: "test",
			},
		},
		&v1.PersistentVolumeClaim{
			ObjectMeta: validMeta("test", "test"),
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: capacity,
					},
				},
			},
		},
		&v1.PersistentVolume{
			ObjectMeta: v1.ObjectMeta{
				Name:   "test",
				Labels: validLabels,
			},
		},
		&v1beta1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
				Labels:    validLabels,
			},
			Spec: v1beta1.DeploymentSpec{
				Replicas: &replicaCount,
				Template: v1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels:    validLabels,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Env: []v1.EnvVar{
									{
										Name:  "test",
										Value: "1",
									},
									{
										Name:  "test2",
										Value: "2",
									},
									{
										Name:  "test3",
										Value: "3",
									},
								},

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
		&v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ts",
				Namespace: "test",
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("test", "test"),
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:     "whatevs",
						Protocol: "TCP",
						Port:     80,
					},
				},
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("test", "test2"),
		},
		&v1beta1.Ingress{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ts",
				Namespace: "test",
			},
		},
		&v1beta1.Ingress{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "www.test.com",
					},
				},
			},
		},
	)
	cluster := Cluster{Interface: client}
	environment, err := cluster.LoadEnvironment("test")
	if err != nil {
		t.Error(err)
	}
	if environment == nil {
		t.Error("Bitesize object is nil")
	}

	if environment.Name != "Development" {
		t.Errorf("Unexpected environment name: %s", environment.Name)
	}

	if len(environment.Services) != 2 {
		t.Errorf("Unexpected service count: %d, expected: 2", len(environment.Services))
	}

	svc := environment.Services[0]
	if svc.Name != "test" {
		t.Errorf("Unexpected service name: %s, expected: test", svc.Name)
	}
	// TODO: test ingresses, env variables, replica count

	if svc.ExternalURL != "www.test.com" {
		t.Errorf("Unexpected external URL: %s, expected: www.test.com", svc.ExternalURL)
	}

	if len(svc.EnvVars) != 3 {
		t.Errorf("Unexpected environment variable count: %d, expected: 3", len(svc.EnvVars))
	}
}
