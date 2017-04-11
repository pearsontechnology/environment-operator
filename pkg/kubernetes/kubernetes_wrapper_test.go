package kubernetes

import (
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/diff"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestKubernetesClusterClient(t *testing.T) {
	// t.Run("service count", testServiceCount)
	// t.Run("volumes", testVolumes)
	t.Run("full bitesize construct", testFullBitesizeEnvironment)
	// t.Run("a/b deployment service", testABSingleService)
}

// func testServiceCount(t *testing.T) {
//
// 	client := fake.NewSimpleClientset(
// 		newService("test", "test1", "test", 80),
// 		newService("test", "test2", "test", 81),
// 	)
//
// 	wrapper := Wrapper{Interface: client}
//
// 	services, err := wrapper.Services("test")
// 	if err != nil {
// 		t.Errorf("Unexpected exception: %s", err.Error())
// 	}
//
// 	if len(services) != 2 {
// 		t.Errorf("Unexpected count of services")
// 	}
// }

// func testVolumes(t *testing.T) {
// 	capacity, _ := resource.ParseQuantity("59G")
// 	validLabels := map[string]string{"creator": "pipeline"}
// 	replicaCount := int32(1)
//
// 	client := fake.NewSimpleClientset(
// 		&v1.PersistentVolumeClaim{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "ts",
// 				Namespace: "test",
// 			},
// 		},
// 		&v1.PersistentVolumeClaim{
// 			ObjectMeta: validMeta("test", "test"),
// 			Spec: v1.PersistentVolumeClaimSpec{
// 				AccessModes: []v1.PersistentVolumeAccessMode{
// 					v1.ReadWriteOnce,
// 				},
// 				Resources: v1.ResourceRequirements{
// 					Requests: v1.ResourceList{
// 						v1.ResourceStorage: capacity,
// 					},
// 				},
// 			},
// 		},
// 		&v1.PersistentVolume{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:   "test",
// 				Labels: validLabels,
// 			},
// 		},
// 		&v1beta1.Deployment{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "test",
// 				Namespace: "test",
// 				Labels:    validLabels,
// 			},
// 			Spec: v1beta1.DeploymentSpec{
// 				Replicas: &replicaCount,
// 				Template: v1.PodTemplateSpec{
// 					ObjectMeta: v1.ObjectMeta{
// 						Name:      "test",
// 						Namespace: "test",
// 						Labels:    validLabels,
// 					},
// 					Spec: v1.PodSpec{
// 						Containers: []v1.Container{
// 							{
// 								VolumeMounts: []v1.VolumeMount{
// 									{
// 										Name:      "test",
// 										MountPath: "/tmp/blah",
// 										ReadOnly:  true,
// 									},
// 								},
// 							},
// 						},
// 						Volumes: []v1.Volume{
// 							{
// 								Name: "test",
// 								VolumeSource: v1.VolumeSource{
// 									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
// 										ClaimName: "test",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		// &api.Service{
// 		// 	ObjectMeta: api.ObjectMeta{
// 		// 		Name:      "ts",
// 		// 		Namespace: "test",
// 		// 	},
// 		// },
// 		// &v1beta1.Ingress{
// 		// 	ObjectMeta: api.ObjectMeta{
// 		// 		Name:      "ts",
// 		// 		Namespace: "test",
// 		// 	},
// 		// },
// 	)
// 	wrapper := Wrapper{Interface: client}
//
// 	// a, err := wrapper.PersistentVolumeClaims("test")
// 	// if err != nil {
// 	// 	t.Errorf("Unexpected exception: %s", err.Error())
// 	// }
// 	//
// 	// if len(a) != 1 {
// 	// 	t.Errorf("Unexpected amount of persistent volume claims returned: %d, expected 1", len(a))
// 	// }
//
// 	// r, err := wrapper.VolumesForDeployment("test", "test")
// 	// if err != nil {
// 	// 	t.Errorf("Error retrieving volumes: %s", err.Error())
// 	// }
// 	//
// 	// if len(r) != 1 {
// 	// 	t.Errorf("Unexpected amount of volumes in VolumesForDeployment: %d", len(r))
// 	// }
// 	//
// 	// if r[0].Size != "59G" {
// 	// 	t.Errorf("Error in volume size, expected: %s, actual: %s", "59G", r[0].Size)
// 	// }
//
// }

// func testABSingleService(t *testing.T) {
// 	capacity, _ := resource.ParseQuantity("59G")
// 	validLabels := map[string]string{"creator": "pipeline"}
// 	replicaCount := int32(1)
//
// 	client := fake.NewSimpleClientset(
// 		&v1.PersistentVolumeClaim{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "ts",
// 				Namespace: "test",
// 			},
// 		},
// 		&v1.PersistentVolumeClaim{
// 			ObjectMeta: validMeta("test", "test"),
// 			Spec: v1.PersistentVolumeClaimSpec{
// 				AccessModes: []v1.PersistentVolumeAccessMode{
// 					v1.ReadWriteOnce,
// 				},
// 				Resources: v1.ResourceRequirements{
// 					Requests: v1.ResourceList{
// 						v1.ResourceStorage: capacity,
// 					},
// 				},
// 			},
// 		},
// 		&v1.PersistentVolume{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:   "test",
// 				Labels: validLabels,
// 			},
// 		},
// 		&v1beta1.Deployment{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "test",
// 				Namespace: "test",
// 				Labels:    validLabels,
// 			},
// 			Spec: v1beta1.DeploymentSpec{
// 				Replicas: &replicaCount,
// 				Template: v1.PodTemplateSpec{
// 					ObjectMeta: v1.ObjectMeta{
// 						Name:      "test",
// 						Namespace: "test",
// 						Labels:    validLabels,
// 					},
// 					Spec: v1.PodSpec{
// 						Containers: []v1.Container{
// 							{
// 								VolumeMounts: []v1.VolumeMount{
// 									{
// 										Name:      "test",
// 										MountPath: "/tmp/blah",
// 										ReadOnly:  true,
// 									},
// 								},
// 							},
// 						},
// 						Volumes: []v1.Volume{
// 							{
// 								Name: "test",
// 								VolumeSource: v1.VolumeSource{
// 									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
// 										ClaimName: "test",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		&v1.Service{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "ts",
// 				Namespace: "test",
// 			},
// 		},
// 		&v1beta1.Ingress{
// 			ObjectMeta: v1.ObjectMeta{
// 				Name:      "ts",
// 				Namespace: "test",
// 			},
// 		},
// 	)
//
// 	// wrapper := Wrapper{Interface: client}
// 	// wrapper.Services("test")
//
// 	// t.Errorf("Not implemented")
// }

func TestApplyEnvironment(t *testing.T) {

	log.SetLevel(log.FatalLevel)
	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: "environment-dev",
				Labels: map[string]string{
					"environment": "environment2",
				},
			},
		},
	)
	cluster := Cluster{Interface: client}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	cluster.ApplyEnvironment(e1)

	e2, err := cluster.LoadEnvironment("environment-dev")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if d := diff.Compare(*e1, *e2); d != "" {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", d)
	}
}

func newDeployment(namespace, name string) *v1beta1.Deployment {
	d := v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{},
		},
	}
	return &d
}

func newService(namespace, serviceName, portName string, portNumber int32) *v1.Service {
	labels := map[string]string{"creator": "pipeline"}

	service := v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "1.1.1.1",
			Ports: []v1.ServicePort{
				{Port: portNumber, Name: portName, Protocol: "TCP"},
			},
		},
	}
	return &service
}

func validMeta(namespace, name string) v1.ObjectMeta {
	validLabels := map[string]string{"creator": "pipeline"}

	if namespace != "" {
		return v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    validLabels,
		}
	}
	return v1.ObjectMeta{
		Name:   name,
		Labels: validLabels,
	}
}
