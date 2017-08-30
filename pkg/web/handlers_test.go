package web

import (
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"testing"
)

func TestPodsReturnForDeployment(t *testing.T) {
	var deployedPods []bitesize.Pod
	var testpod1 = bitesize.Pod{
		Name: "front-pod",
	}

	deployedPods = append(deployedPods, testpod1)

	var testpod2 = bitesize.Pod{
		Name: "back-pod",
	}

	deployedPods = append(deployedPods, testpod2)

	var services = bitesize.Services{
		{Name: "front"},
		{Name: "back"},
		{Name: "podservice"},
	}

	services[2].DeployedPods = deployedPods

	//Test to make sure only the pod for the "front" service is returned
	status := statusForPods(services[0], services[2])
	for _, pod := range status.Pods {
		if pod.Name != "front-pod" {
			t.Errorf("Unexpected Pod returned from podservice. Should be front-pod, but was %v", pod.Name)
		}

	}
	//Test to make sure only the pod for the "back" service is returned
	status = statusForPods(services[1], services[2])
	for _, pod := range status.Pods {
		if pod.Name != "back-pod" {
			t.Errorf("Unexpected Pod returned from podservice. Should be back-pod, but was %v", pod.Name)
		}

	}

}
