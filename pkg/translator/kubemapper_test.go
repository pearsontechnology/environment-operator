package translator

import (
	"os"
	"reflect"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"k8s.io/client-go/pkg/api/v1"
)

func TestThirdPartyResource(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Type = "mysql"
	w.BiteService.Version = "5.6"

	tpr, _ := w.ThirdPartyResource()

	if tpr.Kind != "Mysql" {
		t.Errorf("tpr kind error. Expected: Mysql, got: %s", tpr.Kind)
	}

	if tpr.Spec.Version != w.BiteService.Version {
		t.Errorf("tpr version error: Expected: %s, got: %s", w.BiteService.Version, tpr.Spec.Version)
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

func TestTranslatorHPA(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.HPA.MinReplicas = 1
	w.BiteService.HPA.MaxReplicas = 6
	w.BiteService.HPA.TargetCPUUtilizationPercentage = 51

	h, _ := w.HPA()

	if *h.Spec.MinReplicas != w.BiteService.HPA.MinReplicas {
		t.Errorf("Wrong HPA min replicas value: %+v, expected %+v", *h.Spec.MinReplicas, w.BiteService.HPA.MinReplicas)
	} else if h.Spec.MaxReplicas != w.BiteService.HPA.MaxReplicas {
		t.Errorf("Wrong HPA max replicas value: %+v, expected %+v", h.Spec.MaxReplicas, w.BiteService.HPA.MaxReplicas)
	} else if *h.Spec.TargetCPUUtilizationPercentage != w.BiteService.HPA.TargetCPUUtilizationPercentage {
		t.Errorf("Wrong HPA target CPU utilization percentage value: %+v, expected %+v", *h.Spec.TargetCPUUtilizationPercentage, w.BiteService.HPA.TargetCPUUtilizationPercentage)
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
