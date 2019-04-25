package translator

import (
	"os"
	"reflect"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRD(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Type = "mysql"
	w.BiteService.Version = "5.6"

	tpr, _ := w.CustomResourceDefinition()

	if tpr.Kind != "Mysql" {
		t.Errorf("CRD kind error. Expected: Mysql, got: %s", tpr.Kind)
	}

	if tpr.Spec.Version != w.BiteService.Version {
		t.Errorf("CRD version error: Expected: %s, got: %s", w.BiteService.Version, tpr.Spec.Version)
	}
}

func TestTranslatorIngressLabels(t *testing.T) {
	t.Run("ssl label", testTranslatorIngressSSl)
	t.Run("httpsBackend label", testTranslatorIngressHTTPSBackend)
	t.Run("httpsOnly label", testTranslatorIngressHTTPSOnly)
	t.Run("httpsOnly label", testTranslatorIngressHTTP2)
}

func testTranslatorIngressSSl(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Ssl = "true"
	w.BiteService.ExternalURL = []string{"www.test.com"}

	ingress, _ := w.Ingress()

	if ingress.Labels["ssl"] != "true" {
		t.Errorf("Unexpected ingress ssl value: %+v", ingress.Labels["ssl"])
	}
}

func testTranslatorIngressHTTP2(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.HTTP2 = "true"
	w.BiteService.ExternalURL = []string{"www.test.com"}

	ingress, _ := w.Ingress()

	if ingress.Labels["http2"] != "true" {
		t.Errorf("Unexpected ingress http2 value: %+v", ingress.Labels["http2"])
	}
}

func TestDockerPullSecrets(t *testing.T) {
	w := BuildKubeMapper()
	os.Setenv("DOCKER_PULL_SECRETS", "pullsecret")
	deploy, _ := w.Deployment()
	os.Unsetenv("DOCKER_PULL_SECRETS")
	var testValue []v1.LocalObjectReference
	testValue = []v1.LocalObjectReference{{Name: "pullsecret"}}
	for i := range testValue {
		var deployImagePullSecret []v1.LocalObjectReference
		deployImagePullSecret = deploy.Spec.Template.Spec.ImagePullSecrets
		if testValue[i] != deployImagePullSecret[i] {
			t.Errorf("Unexpected Value for ImagePullSecret. Expected= %+v Actual= %+v", testValue[i], deployImagePullSecret[i])
		}
	}
}

func TestVolumeFromSecret(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Name = "test"
	w.Namespace = "test"
	w.BiteService.Volumes = []bitesize.Volume{
		{Name: "internal-ca", Path: "/tmp/vol1", Type: "secret"},
	}
	generatedVolumes, _ := w.volumes()

	v := w.BiteService.Volumes

	expectedVolumes := []v1.Volume{
		{
			Name: v[0].Name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: v[0].Name,
				},
			},
		},
	}

	if !reflect.DeepEqual(generatedVolumes, expectedVolumes) {
		t.Errorf("incorrect volumes: %v generated; expecting: %v ", generatedVolumes, expectedVolumes)
	}
}

func testTranslatorIngressHTTPSBackend(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.HTTPSBackend = "true"
	w.BiteService.ExternalURL = []string{"www.test.com"}

	ingress, _ := w.Ingress()

	if ingress.Labels["httpsBackend"] != "true" {
		t.Errorf("Unexpected ingress httpsBackend value: %+v", ingress.Labels["httpsBackend"])
	}
}

func testTranslatorIngressHTTPSOnly(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.HTTPSOnly = "true"
	w.BiteService.ExternalURL = []string{"www.test.com"}

	ingress, _ := w.Ingress()

	if ingress.Labels["httpsOnly"] != "true" {
		t.Errorf("Unexpected ingress httpsOnly value: %+v", ingress.Labels["httpsOnly"])
	}
}

func TestTranslatorIngressBackendOverride(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.ExternalURL = []string{"www.test.com"}
	w.BiteService.Backend = "www.example.com"

	ingress, _ := w.Ingress()
	result := ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName

	if result != "www.example.com" {
		t.Errorf("wrong ingress backend value: %s, expecting: %s", result, w.BiteService.Backend)
	}
}

func TestTranslatorIngressBackendPortOverride(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.ExternalURL = []string{"www.test.com"}
	w.BiteService.Backend = "www.example.com"
	w.BiteService.BackendPort = 81

	ingress, _ := w.Ingress()
	result := int(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal)

	if result != w.BiteService.BackendPort {
		t.Errorf("wrong ingress backend_port value: %v, expecting: %v", result, w.BiteService.BackendPort)
	}
}

func BuildKubeMapper() *KubeMapper {
	m := &KubeMapper{
		BiteService: &bitesize.Service{
			Name:  "test",
			Ports: []int{80},
		},
		Namespace: "testns",
	}
	m.Config.Project = "project"
	m.Config.DockerRegistry = "registry"
	return m
}

func TestInitContainers(t *testing.T) {
	w := BuildKubeMapper()

	w.BiteService.InitContainers = &[]bitesize.Container{
		bitesize.Container{
			Name:        "nginx",
			Version:     "1.15.10-1~stretch",
			Application: "nginx_init",
			EnvVars: []bitesize.EnvVar{
				{Name: "test1", Value: "test1"},
				{Name: "testpodfield", PodField: "metadata.namespace"},
			},
		},
	}

	d, _ := w.Deployment()

	if d.Spec.Template.Spec.InitContainers[0].Name != "nginx" {
		t.Errorf("Wrong name for the init container %s", d.Spec.Template.Spec.InitContainers[0].Image)
	}

	if d.Spec.Template.Spec.InitContainers[0].Env[0].Name != "test1" {
		t.Errorf("Wrong name for the init container env %s", d.Spec.Template.Spec.InitContainers[0].Env[0].Name)
	}
}

func TestTranslatorHPA(t *testing.T) {

	w := BuildKubeMapper()
	w.BiteService.HPA.MinReplicas = 1
	w.BiteService.HPA.MaxReplicas = 6
	w.BiteService.HPA.Metric.TargetAverageUtilization = 51
	w.BiteService.HPA.Metric.Name = "cpu"

	h, _ := w.HPA()

	if *h.Spec.MinReplicas != w.BiteService.HPA.MinReplicas {
		t.Errorf("Wrong HPA min replicas value: %+v, expected %+v", *h.Spec.MinReplicas, w.BiteService.HPA.MinReplicas)
	} else if h.Spec.MaxReplicas != w.BiteService.HPA.MaxReplicas {
		t.Errorf("Wrong HPA max replicas value: %+v, expected %+v", h.Spec.MaxReplicas, w.BiteService.HPA.MaxReplicas)
	} else if *h.Spec.Metrics[0].Resource.TargetAverageUtilization != w.BiteService.HPA.Metric.TargetAverageUtilization {
		t.Errorf("Wrong HPA CPU target Average value: %+v, expected %+v", *h.Spec.Metrics[0].Resource.TargetAverageUtilization, w.BiteService.HPA.Metric.TargetAverageUtilization)
	}

	w.BiteService.HPA.Metric.TargetAverageValue = "10m"
	w.BiteService.HPA.Metric.Name = "custom_metric"
	h, _ = w.HPA()
	targetAverageValue, _ := resource.ParseQuantity(w.BiteService.HPA.Metric.TargetAverageValue)

	if h.Spec.Metrics[0].Pods.TargetAverageValue != targetAverageValue {
		t.Errorf("Wrong HPA CPU target Average value: %+v, expected %+v", h.Spec.Metrics[0].Pods.TargetAverageValue, w.BiteService.HPA.Metric.TargetAverageValue)
	}
}

func TestTranslatorEnvVars(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Replicas = 1
	w.BiteService.Name = "test"
	w.BiteService.Application = "test"
	w.BiteService.Version = "test"
	w.BiteService.EnvVars = []bitesize.EnvVar{
		{Name: "test1", Value: "test1"},
		{Name: "testpodfield", PodField: "metadata.namespace"},
	}

	d, _ := w.Deployment()

	generatedEnvVars := d.Spec.Template.Spec.Containers[0].Env
	expectedEnvVars := []v1.EnvVar{
		{Name: "test1", Value: "test1"},
		{
			Name: "testpodfield",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	if !reflect.DeepEqual(generatedEnvVars, expectedEnvVars) {
		t.Errorf("incorrect environment variables: %v generated; expecting: %v ", generatedEnvVars, expectedEnvVars)
	}
}

func TestTranslatorPVCs(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Name = "test"
	w.Namespace = "test"
	w.BiteService.Volumes = []bitesize.Volume{
		{Name: "vol1", Path: "/tmp/vol1", Modes: "ReadWriteOnce", Size: "1Gi", Type: "EFS"},
		{Name: "vol2", Path: "/tmp/vol2", Modes: "ReadOnlyMany", Size: "1Gi", Type: "eBs"},
	}

	generatedPVCs, _ := w.PersistentVolumeClaims()
	expectedPVCs := []v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vol1",
				Namespace: "test",
				Labels: map[string]string{
					"creator":    "pipeline",
					"deployment": "test",
					"mount_path": "2Ftmp2Fvol1",
					"size":       "1Gi",
					"type":       "efs",
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": "aws-efs",
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: getAccessModesFromString("ReadWriteOnce"),
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vol2",
				Namespace: "test",
				Labels: map[string]string{
					"creator":    "pipeline",
					"deployment": "test",
					"mount_path": "2Ftmp2Fvol2",
					"size":       "1Gi",
					"type":       "ebs",
				},
				Annotations: map[string]string{
					"volume.beta.kubernetes.io/storage-class": "aws-ebs",
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: getAccessModesFromString("ReadOnlyMany"),
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(generatedPVCs, expectedPVCs) {
		t.Errorf("incorrect PVCs: %v generated; expecting: %v ", generatedPVCs, expectedPVCs)
	}

}
