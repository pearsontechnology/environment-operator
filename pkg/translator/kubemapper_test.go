package translator

import (
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"testing"
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
}

func testTranslatorIngressSSl(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.Ssl = "true"

	ingress, _ := w.Ingress()

	if ingress.Labels["ssl"] != "true" {
		t.Errorf("Unexpected ingress ssl value: %+v", ingress.Labels["ssl"])
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

	ingress, _ := w.Ingress()

	if ingress.Labels["httpsBackend"] != "true" {
		t.Errorf("Unexpected ingress httpsBackend value: %+v", ingress.Labels["httpsBackend"])
	}
}

func testTranslatorIngressHTTPSOnly(t *testing.T) {
	w := BuildKubeMapper()
	w.BiteService.HTTPSOnly = "true"

	ingress, _ := w.Ingress()

	if ingress.Labels["httpsOnly"] != "true" {
		t.Errorf("Unexpected ingress httpsOnly value: %+v", ingress.Labels["httpsOnly"])
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
