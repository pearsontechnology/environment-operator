package cluster

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddDeploymentSetting(t *testing.T) {
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "sample",
			Annotations: map[string]string{
				"deployment_method": "bluegreen",
				"deployment_active": "blue",
			},
		},
	}
	serviceMap := &ServiceMap{}

	serviceMap.AddService(svc)

	biteservice := serviceMap.CreateOrGet("test")
	if biteservice.DeploymentMethod() != "bluegreen" {
		t.Errorf("unexpected deployment method. expected bluegreen, got %+v", biteservice.DeploymentMethod())
	}

	if biteservice.ActiveDeploymentName() != "test-blue" {
		t.Errorf("unexpected active deployment name. expected test-blue, got: %+v", biteservice.ActiveDeploymentName())
	}
}
