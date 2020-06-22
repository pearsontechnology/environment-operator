package cluster

import (
	"fmt"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/diff"
	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	fakecrd "github.com/pearsontechnology/environment-operator/pkg/util/k8s/fake"
	apps_v1 "k8s.io/api/apps/v1"
	autoscale_v2beta2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	netwk_v1beta1 "k8s.io/api/networking/v1beta1"
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
	runningCluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	envFromConfigFile, err := bitesize.LoadEnvironment(
		"../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	runningCluster.ApplyIfChanged(envFromConfigFile)

	// load an environment config over top of currently-running config
	envToCompare, err := runningCluster.ScrapeResourcesForNamespace("environment-dev")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	/* Uncomment this if you want to debug the YAML marshaling
	cfgYAMLEnvToCompare, err := yaml.Marshal(&envToCompare)
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}
	t.Errorf("envToCompare environment-dev: %s", string(cfgYAMLEnvToCompare))

	cfgYAMLEnvFromConfigFile, err := yaml.Marshal(&envFromConfigFile)
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}
	t.Errorf("envFromConfigFile environment2: %s", string(cfgYAMLEnvFromConfigFile))
	*/

	// There shouldn't be changes b/t runningCluster + envToCompare
	// (they will be synced with the apply below)
	// current is envFromConfigFile
	// new is envToCompare
	runningCluster.ApplyEnvironment(envFromConfigFile, envToCompare)

	// envFromConfigFile is desired config.
	// *envToCompare is existing config
	if diff.Compare(*envFromConfigFile, *envToCompare) {
		t.Errorf("ApplyEnvironment: Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}

	// environments2.bitesize removes annotated_service2 and testdb from environment2
	// The diff b/t envToCompare + env2ToCompare should only be the testdb change
	// (b/c annotated_service2 doesn't have a "version" field in source config)
	env2ToCompare, err := bitesize.LoadEnvironment(
		"../../test/assets/environments2.bitesize", "environment2")

	if !diff.Compare(*envToCompare, *env2ToCompare) {
		fmt.Printf("%+v\n", diff.Changes())
		t.Errorf("expected diff, got none")
	}
	_, exists := diff.Changes()["testdb"]
	if !exists {
		t.Errorf("Expected testdb to exist in the diff, yet it does not exist: %s", diff.Changes())
	}
}

func TestApplyIfChanged(t *testing.T) {

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
	runningCluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	envFromConfigFile, err := bitesize.LoadEnvironment(
		"../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	// initialize running cluster
	runningCluster.ApplyIfChanged(envFromConfigFile)
}

func TestShouldDeployOnChange(t *testing.T) {

	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	// Mark all services in the initial environment as deployed
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

func newDeployment(namespace, name string) *apps_v1.Deployment {
	d := apps_v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "1",
			},
		},
		Spec: apps_v1.DeploymentSpec{
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
		"prsn.io",
		"v1",
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

func loadTestExternalSecrets() *fakerest.RESTClient {
	return fakecrd.CRDClient(
		"kubernetes-client.io",
		"v1",
		&ext.ExternalSecret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ExternalSecret",
				APIVersion: "kubernetes-client.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "testexternalsecret",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			SecretDescriptor: ext.ExternalSecretSecretDescriptor{
				BackendType: "secretsManager",
				Compressed:  true,
				Type:        "kubernetes.io/tls",
				Data: []map[string]string{
					{
						"key":  fmt.Sprintf("tls/test.crt"),
						"name": "tls.crt",
					},
					{
						"key":  fmt.Sprintf("tls/test.key"),
						"name": "tls.key",
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
		&apps_v1.Deployment{
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
			Status: apps_v1.DeploymentStatus{
				AvailableReplicas: 1,
				Replicas:          1,
				UpdatedReplicas:   1,
			},
			Spec: apps_v1.DeploymentSpec{
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
										Name: "TEST4",
										ValueFrom: &v1.EnvVarSource{
											SecretKeyRef: &v1.SecretKeySelector{
												LocalObjectReference: v1.LocalObjectReference{Name: "test4"},
												Key:                  "ttt",
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
		&netwk_v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ts",
				Namespace: "test",
			},
		},
		&netwk_v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
				Labels: map[string]string{
					"creator": "pipeline",
				},
			},
			Spec: netwk_v1beta1.IngressSpec{
				Rules: []netwk_v1beta1.IngressRule{
					{
						Host: "www.test.com",
						IngressRuleValue: netwk_v1beta1.IngressRuleValue{
							HTTP: &netwk_v1beta1.HTTPIngressRuleValue{
								Paths: []netwk_v1beta1.HTTPIngressPath{
									{
										Path: "/",
										Backend: netwk_v1beta1.IngressBackend{
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
	environment, err := cluster.ScrapeResourcesForNamespace("test")
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
	environment, err := cluster.ScrapeResourcesForNamespace("test")
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

	if secretEnvVar.Secret != "TEST4" || secretEnvVar.Value != "test4/ttt" {
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
	)

	cluster := Cluster{
		Interface: client,
		CRDClient: crdcli,
	}

	config.Env.UseAuth = false // setting auth disabled
	e1, err := bitesize.LoadEnvironment("../../test/assets/environments.bitesize", "environment13")
	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	e2, err := cluster.ScrapeResourcesForNamespace("environment-dev")

	_ = cluster.ApplyEnvironment(e1, e2)

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
	environment, _ := cluster.ScrapeResourcesForNamespace("test")
	testService := environment.Services.FindByName("test")

	if testService.Annotations["existing_annotation"] != "exist" {
		t.Error("Existing annotation is not loaded from the cluster before apply")
	}

	e1, _ := bitesize.LoadEnvironment("../../test/assets/annotations.bitesize", "test")
	_ = cluster.ApplyEnvironment(e1, e1)

	e2, _ := cluster.ScrapeResourcesForNamespace("test")
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

	_ = cluster.ApplyEnvironment(e1, e1)

	e2, err := cluster.ScrapeResourcesForNamespace("environment-dev")

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
		&autoscale_v2beta2.HorizontalPodAutoscaler{
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
			Spec: autoscale_v2beta2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscale_v2beta2.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       "hpa-service",
					APIVersion: "apps/v1",
				},
				MinReplicas: &min,
				MaxReplicas: 5,
				Metrics: []autoscale_v2beta2.MetricSpec{
					{
						Type: autoscale_v2beta2.ObjectMetricSourceType,
						Object: &autoscale_v2beta2.ObjectMetricSource{
							Target: autoscale_v2beta2.MetricTarget{
								Type:         "AverageVale",
								AverageValue: &customMetricValue,
							},
							Metric: autoscale_v2beta2.MetricIdentifier{
								Name: "custom_metric",
							},
						},
					},
					{
						Type: autoscale_v2beta2.ResourceMetricSourceType,
						Resource: &autoscale_v2beta2.ResourceMetricSource{
							Target: autoscale_v2beta2.MetricTarget{
								Type:               "Utilization",
								AverageUtilization: &target,
							},
							Name: "cpu",
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

	_ = cluster.ApplyEnvironment(e1, e1)

	e2, err := cluster.ScrapeResourcesForNamespace("environment-dev")

	if err != nil {
		t.Fatalf("Unexpected err: %s", err.Error())
	}

	if diff.Compare(*e1, *e2) {
		t.Errorf("Expected loaded environments to be equal, yet diff is: %s", diff.Changes())
	}
}
func loadEmptyCRDs() *fakerest.RESTClient {
	return fakecrd.CRDClient("prsn.io", "v1")
}
