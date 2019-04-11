package cluster

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	fakecrd "github.com/pearsontechnology/environment-operator/pkg/util/k8s/fake"
	v1beta2_apps "k8s.io/api/apps/v1beta2"
	autoscale_v2beta1 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	v1beta1_ext "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	fakerest "k8s.io/client-go/rest/fake"
)

// func init() {
// 	// Let our fake server handle our tprs
// 	// by registering prsn.io/v1 resources
// 	// it's easier in client-go v1.6
// 	m := registered.DefaultAPIRegistrationManager

// 	groupversion := runtime.GroupVersion{
// 		Group:   "prsn.io",
// 		Version: "v1",
// 	}
// 	groupversions := []runtime.GroupVersion{groupversion}
// 	groupmeta := apimachinery.GroupMeta{
// 		GroupVersion: groupversion,
// 	}

// 	m.RegisterVersions(groupversions)
// 	m.AddThirdPartyAPIGroupVersions(groupversion)
// 	m.RegisterGroup(groupmeta)
// }

func TestKubernetesClusterClient(t *testing.T) {
	// t.Run("service count", testServiceCount)
	// t.Run("volumes", testVolumes)
	t.Run("full bitesize construct", testFullBitesizeEnvironment)
	t.Run("test service ports", testServicePorts)
	// t.Run("a/b deployment service", testABSingleService)
}

func TestApplyEnvironment(t *testing.T) {

	log.SetLevel(log.FatalLevel)
	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "environment-dev",
				Labels: map[string]string{
					"environment": "environment2",
				},
			},
		},
	)
	crdcli := loadTestCRDs()
	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	cluster.ApplyIfChanged(e1)

	e2, err := cluster.LoadEnvironment("environment-dev")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	//There should be no changes between environments e1 and e2 (they will be synced with the apply below)
	//	diff.Compare(*e1, *e2)
	cluster.ApplyEnvironment(e1, e2)
	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}

	//environments2.bitesize removes annotated_service2 and testdb from environment2
	//The diff between e2 and e3 should only contain the testdb change as annotated_service2 didn't have a "version" field in the source config
	e3, err := bitesize.LoadEnvironment("../../test/assets/environments2.bitesize", "environment2")
	if !diff.Compare(*e2, *e3) {
		fmt.Printf("%+v\n", diff.Changes())
		t.Errorf("expected diff, got none")
	}
	_, exists := diff.Changes()["testdb"]
	if !exists {
		t.Errorf("Expected testdb to exist in the diff, yet it does not exist: %s", diff.Changes())
	}
}

func TestShouldDeployOnChange(t *testing.T) {

	log.SetLevel(log.FatalLevel)

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	//Mark all services in the initial environment as deployed
	for i := range e1.Services {
		e1.Services[i].Status.DeployedAt = "Current Time"
	}

	e2, err := bitesize.LoadEnvironment("../../test/assets/environments2.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	diff.Compare(*e1, *e2)

	deploy := shouldDeployOnChange(e1, e2, "annotated_service2")

	if deploy {
		t.Error("Expected that the annotated_service2 service should not be marked for deploy, but it was.")
	}

	deploy = shouldDeployOnChange(e1, e2, "testdb")

	if !deploy {
		t.Error("Expected that the testdb service should be marked for deploy, but it was not.")
	}

}

/*

//Currently disabled due to https://github.com/kubernetes/client-go/issues/196  .  Unable to mock out Request Stream() request
that is made when Pod logs are retrieved by the LoadPods() function.

func TestGetPods(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	labels := map[string]string{"creator": "pipeline"}
	client := fake.NewSimpleClientset(
		&v1.Pod{
			TypeMeta: runtime.TypeMeta{
				Kind:       "pod",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "front",
				Namespace: "dev",
				Labels:    labels,
			},
		},
	)
	crdclient := loadTestCRDs()
	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}
	pods, err := cluster.LoadPods("dev")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if !strings.Contains(pods[0].Name, "front") {
		t.Errorf("Expected 'front' pod to be retrieved")
	}

}
*/

func newDeployment(namespace, name string) *v1beta1_ext.Deployment {
	d := v1beta1_ext.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "1",
			},
		},
		Spec: v1beta1_ext.DeploymentSpec{
			Template: v1.PodTemplateSpec{},
		},
	}
	return &d
}

func newService(namespace, serviceName, portName string, portNumber int32) *v1.Service {
	labels := map[string]string{"creator": "pipeline"}

	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
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

func validMeta(namespace, name string) metav1.ObjectMeta {
	validLabels := map[string]string{"creator": "pipeline"}

	if namespace != "" {
		return metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    validLabels,
		}
	}
	return metav1.ObjectMeta{
		Name:   name,
		Labels: validLabels,
	}
}

func loadTestCRDs() *fakerest.RESTClient {
	return fakecrd.CRDClient(
		&ext.PrsnExternalResource{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Neptune",
				APIVersion: "prsn.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "testdb",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: ext.PrsnExternalResourceSpec{
				Version: "1.0.1.0.200264.0",
				Options: map[string]interface{}{
					"ApplyImmediately": "true",
					"db_instances": []map[string]string{
						map[string]string{
							"db_name":           "db01",
							"db_instance_class": "db.r4.2xlarge",
						},
						map[string]string{
							"db_name":           "db02",
							"db_instance_class": "db.r4.xlarge",
						},
					},
				},
			},
		},
	)
}

func loadTestEnvironment() *fake.Clientset {
	capacity, _ := resource.ParseQuantity("59G")
	validLabels := map[string]string{"creator": "pipeline"}
	nsLabels := map[string]string{"environment": "Development"}
	replicaCount := int32(1)
	cpulimit, _ := resource.ParseQuantity("1000m")
	memlimit, _ := resource.ParseQuantity("500Mi")
	cpurequest, _ := resource.ParseQuantity("500m")
	memrequest, _ := resource.ParseQuantity("200Mi")

	return fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test",
				Labels: nsLabels,
			},
		},
		&v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
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
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test",
				Labels: validLabels,
			},
		},
		&v1beta1_ext.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
				Labels: map[string]string{
					"creator":     "pipeline",
					"name":        "hpaservice",
					"application": "some-app",
					"version":     "some-version",
				},
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
			Status: v1beta1_ext.DeploymentStatus{
				AvailableReplicas: 1,
				Replicas:          1,
				UpdatedReplicas:   1,
			},
			Spec: v1beta1_ext.DeploymentSpec{
				Replicas: &replicaCount,
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels:    validLabels,
						Annotations: map[string]string{
							"existing_annotation": "exist",
						},
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
									{
										Name: "test4",
										ValueFrom: &v1.EnvVarSource{
											SecretKeyRef: &v1.SecretKeySelector{
												Key: "ttt",
											},
										},
									},
								},
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "test",
										MountPath: "/tmp/blah",
										ReadOnly:  true,
									},
								},
								Command: []string{
									"test1",
									"test2",
								},
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										"cpu":    cpulimit,
										"memory": memlimit,
									},
									Requests: v1.ResourceList{
										"cpu":    cpurequest,
										"memory": memrequest,
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
			ObjectMeta: metav1.ObjectMeta{
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
					{
						Name:     "whatevs2",
						Protocol: "TCP",
						Port:     8081,
					},
				},
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("test", "test2"),
		},
		&v1beta1_ext.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ts",
				Namespace: "test",
			},
		},
		&v1beta1_ext.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: v1beta1_ext.IngressSpec{
				Rules: []v1beta1_ext.IngressRule{
					{
						Host: "www.test.com",
						IngressRuleValue: v1beta1_ext.IngressRuleValue{
							HTTP: &v1beta1_ext.HTTPIngressRuleValue{
								Paths: []v1beta1_ext.HTTPIngressPath{
									{
										Path: "/",
										Backend: v1beta1_ext.IngressBackend{
											ServiceName: "test",
										},
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

func testServicePorts(t *testing.T) {
	client := loadTestEnvironment()
	crdcli := loadTestCRDs()
	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}
	environment, err := cluster.LoadEnvironment("test")
	if err != nil {
		t.Error(err)
	}

	svc := environment.Services.FindByName("test")
	if !util.EqualArrays(svc.Ports, []int{80, 8081}) {
		t.Errorf("Ports not equal. Expected: [80 8081], got: %v", svc.Ports)
	}
}

func testFullBitesizeEnvironment(t *testing.T) {

	client := loadTestEnvironment()
	crdcli := loadTestCRDs()
	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}
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

	if len(environment.Services) != 3 {
		t.Errorf("Unexpected service count: %d, expected: 3", len(environment.Services))
	}

	svc := environment.Services[0]
	if svc.Name != "test" {
		t.Errorf("Unexpected service name: %s, expected: test", svc.Name)
	}
	// TODO: test ingresses, env variables, replica count

	if len(svc.ExternalURL) != 0 {
		if svc.ExternalURL[0] != "www.test.com" {
			t.Errorf("Unexpected external URL: %s, expected: www.test.com", svc.ExternalURL)
		}
	}

	if len(environment.Services[0].Commands) != 2 {
		t.Errorf("Unexpected Commands: %s, expected: test1,test2", environment.Services[0].Commands)
	}

	if len(svc.EnvVars) != 4 {
		t.Errorf("Unexpected environment variable count: %d, expected: 4", len(svc.EnvVars))
	}

	secretEnvVar := svc.EnvVars[3]

	if secretEnvVar.Secret != "test4" || secretEnvVar.Value != "ttt" {
		t.Errorf("Unexpected envvar[3]: %+v", secretEnvVar)
	}
}

func TestInitContainers(t *testing.T) {
	crdcli := loadEmptyCRDs()
	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "environment-dev",
				Labels: map[string]string{
					"environment": "environment-dev",
				},
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("environment-dev", "init_svc"),
		},
	)

	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment13")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	e2, err := cluster.LoadEnvironment("environment-dev")

	cluster.ApplyEnvironment(e2, e1)

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}
}

func TestEnvironmentAnnotations(t *testing.T) {
	client := loadTestEnvironment()
	crdcli := loadTestCRDs()
	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}
	environment, _ := cluster.LoadEnvironment("test")
	testService := environment.Services.FindByName("test")

	if testService.Annotations["existing_annotation"] != "exist" {
		t.Error("Existing annotation is not loaded from the cluster before apply")
	}

	e1, _ := bitesize.LoadEnvironment("../../test/assets/annotations.bitesize", "test")
	cluster.ApplyEnvironment(e1, e1)

	e2, _ := cluster.LoadEnvironment("test")
	testService = e2.Services.FindByName("test")

	if testService.Annotations["existing_annotation"] != "exist" {
		t.Error("Existing annotation is not loaded from the cluster after apply")
	}

}

func TestApplyNewHPA(t *testing.T) {

	crdcli := loadEmptyCRDs()
	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "environment-dev",
				Labels: map[string]string{
					"environment": "environment-dev",
				},
			},
		},
	)

	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment3")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	cluster.ApplyEnvironment(e1, e1)

	e2, err := cluster.LoadEnvironment("environment-dev")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}
}

func TestApplyExistingHPA(t *testing.T) {
	var min, target int32 = 2, 75
	customMetricValue, _ := resource.ParseQuantity("200")
	crdcli := loadEmptyCRDs()
	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "environment-dev",
				Labels: map[string]string{
					"environment": "environment-dev",
				},
			},
		},
		&autoscale_v2beta1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpa-service",
				Namespace: "environment-dev",
				Labels: map[string]string{
					"creator":     "pipeline",
					"name":        "hpa-service",
					"application": "some-app",
					"version":     "some-version",
				},
			},
			Spec: autoscale_v2beta1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscale_v2beta1.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       "hpa-service",
					APIVersion: "extensions/v1beta1",
				},
				MinReplicas: &min,
				MaxReplicas: 5,
				Metrics: []autoscale_v2beta1.MetricSpec{
					{
						Type: autoscale_v2beta1.ObjectMetricSourceType,
						Object: &autoscale_v2beta1.ObjectMetricSource{
							TargetValue: customMetricValue,
							MetricName:  "custom_metric",
						},
					},
					{
						Type: autoscale_v2beta1.ResourceMetricSourceType,
						Resource: &autoscale_v2beta1.ResourceMetricSource{
							TargetAverageUtilization: &target,
							Name:                     "memory",
						},
					},
				},
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("environment-dev", "hpa-service"),
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
	)

	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment3")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	cluster.ApplyEnvironment(e1, e1)

	e2, err := cluster.LoadEnvironment("environment-dev")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}
}
func loadEmptyCRDs() *fakerest.RESTClient {
	return fakecrd.CRDClient()
}

func TestApplyMongoStatefulSet(t *testing.T) {
	crdcli := loadEmptyCRDs()
	cpulimit, _ := resource.ParseQuantity(config.Env.LimitDefaultCPU)
	memlimit, _ := resource.ParseQuantity(config.Env.LimitDefaultMemory)
	cpurequests, _ := resource.ParseQuantity(config.Env.RequestsDefaultCPU)

	client := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "environment-mongo",
				Labels: map[string]string{
					"environment": "environment-mongo",
				},
			},
		},
		&v1beta2_apps.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mongo",
				Namespace: "environment-mongo",
				Labels: map[string]string{
					"creator":     "pipeline",
					"name":        "hpaservice",
					"application": "some-app",
					"version":     "3.4",
				},
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
			Status: v1beta2_apps.StatefulSetStatus{
				Replicas: 1,
			},
			Spec: v1beta2_apps.StatefulSetSpec{
				ServiceName: "mongo",
				Replicas:    &[]int32{3}[0],
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mongo",
						Namespace: "test",
						Labels: map[string]string{
							"creator":     "pipeline",
							"name":        "mongo",
							"application": "some-app",
							"version":     "3.4",
							"role":        "mongo",
						},
						Annotations: map[string]string{
							"existing_annotation": "exist",
						},
					},
					Spec: v1.PodSpec{
						TerminationGracePeriodSeconds: &[]int64{10}[0],
						Containers: []v1.Container{
							{
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "mongo-persistent-storage",
										MountPath: "/data/db",
										ReadOnly:  true,
									},
									{
										Name:      "secrets-volume",
										MountPath: "/etc/secrets-volume",
										ReadOnly:  true,
									},
								},
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										"cpu": cpurequests,
									},
									Limits: v1.ResourceList{
										"cpu":    cpulimit,
										"memory": memlimit,
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
							{
								Name: "secrets-volume",
								VolumeSource: v1.VolumeSource{
									Secret: &v1.SecretVolumeSource{
										SecretName:  "shared-bootstrap-data",
										DefaultMode: &[]int32{256}[0],
									},
								},
							},
						},
					},
				},
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mongo-persistent-storage",
							Namespace: "environment-mongo",
							Annotations: map[string]string{
								"volume.beta.kubernetes.io/storage-class": "aws-ebs",
							},
							Labels: map[string]string{
								"creator":    "pipeline",
								"deployment": "mongo",
								"mount_path": "/data/db",
								"size":       "10G",
								"type":       "ebs",
							},
						},
						Spec: v1.PersistentVolumeClaimSpec{
							AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceName(v1.ResourceStorage): resource.MustParse("10G"),
								},
							},
						},
					},
				},
			},
		},
		&v1.Service{
			ObjectMeta: validMeta("environment-mongo", "mongo"),
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:     "tcp-port-27017",
						Protocol: "TCP",
						Port:     27017,
					},
				},
			},
		},
	)

	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment4")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	cluster.ApplyEnvironment(e1, e1)

	e2, err := cluster.LoadEnvironment("environment-mongo")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}
}
