package k8s

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	autoscale_v1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
)

func TestHPACreate(t *testing.T) {
	min := new(int32)
	target := new(int32)
	*min = 1
	*target = 50
	fakeHPAClient := createFakeHPAClient()
	newHPA := &autoscale_v1.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:      "newhpa",
			Namespace: "sample",
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        "newhpa",
				"application": "myapp",
				"version":     "0.1",
			},
		},
		Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "newdeployment",
				APIVersion: "v1beta1",
			},
			MinReplicas:                    min,
			MaxReplicas:                    int32(3),
			TargetCPUUtilizationPercentage: target,
		},
	}

	if err := fakeHPAClient.Apply(newHPA); err != nil {
		t.Errorf("Error creating new hpa %s", err.Error())
	}

	if _, err := fakeHPAClient.Get("newhpa"); err != nil {
		t.Errorf("Error getting newly created hpa %s", err.Error())
	}

}

func TestHPAUpdate(t *testing.T) {
	var min, target int32 = 1, 50
	fakeHPAClient := createFakeHPAClient()
	updatedHPA := &autoscale_v1.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fakehpa",
			Namespace: "sample",
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        "fakehpa",
				"application": "updatedmyapp",
				"version":     "0.2",
			},
		},
		Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "newdeployment",
				APIVersion: "v1beta1",
			},
			MinReplicas:                    &min,
			MaxReplicas:                    int32(3),
			TargetCPUUtilizationPercentage: &target,
		},
	}
	if err := fakeHPAClient.Update(updatedHPA); err != nil {
		t.Errorf("Error updating hpa %s", err.Error())
	}

	hpa, err := fakeHPAClient.Get("fakehpa")
	if err != nil {
		t.Errorf("Error getting updated hpa %s", err.Error())
	}
	if hpa.ObjectMeta.Labels["application"] != "updatedmyapp" {
		t.Error("HPA was not updated succesfully")
	}

}

func TestHPAApplyNew(t *testing.T) {
	var min, target int32 = 2, 75
	fakeHPAClient := createFakeHPAClient()
	newHPA := &autoscale_v1.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:      "newhpa",
			Namespace: "sample",
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        "newhpa",
				"application": "myapp",
				"version":     "0.1",
			},
		},
		Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "newdeployment",
				APIVersion: "v1beta1",
			},
			MinReplicas:                    &min,
			MaxReplicas:                    int32(3),
			TargetCPUUtilizationPercentage: &target,
		},
	}
	if err := fakeHPAClient.Apply(newHPA); err != nil {
		t.Errorf("Error applying new hpa %s", err.Error())
	}

	_, err := fakeHPAClient.Get("newhpa")
	if err != nil {
		t.Errorf("Error getting  newhpa %s", err.Error())
	}
}

func TestHPAApplyExisting(t *testing.T) {
	var min, target int32 = 1, 50
	fakeHPAClient := createFakeHPAClient()
	newHPA := &autoscale_v1.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fakehpa",
			Namespace: "sample",
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        "fakehpa",
				"application": "updatedmyapp",
				"version":     "0.1",
			},
		},
		Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "newdeployment",
				APIVersion: "v1beta1",
			},
			MinReplicas:                    &min,
			MaxReplicas:                    int32(3),
			TargetCPUUtilizationPercentage: &target,
		},
	}
	if err := fakeHPAClient.Apply(newHPA); err != nil {
		t.Errorf("Error applying existing hpa %s", err.Error())
	}

	hpa, err := fakeHPAClient.Get("fakehpa")
	if err != nil {
		t.Errorf("Error getting  fakehpa %s", err.Error())
	}
	if hpa.ObjectMeta.Labels["application"] != "updatedmyapp" {
		t.Error("Existing HPA apply was not succesfull.")
	}

}

func TestHPADestroy(t *testing.T) {
	fakeHPAClient := createFakeHPAClient()
	err := fakeHPAClient.Destroy("fakehpa")
	if err != nil {
		t.Errorf("Error destroying hpa %s", err.Error())
	}
	if fakeHPAClient.Exist("fakehpa") {
		t.Error("Hpa was not destroyed")
	}
}

func TestHPAList(t *testing.T) {
	fakeHPAClient := createFakeHPAClient()

	hpaSlice, err := fakeHPAClient.List()
	if err != nil {
		t.Errorf("Error retrieving hpa list %s", err.Error())
	}
	if len(hpaSlice) == 0 {
		t.Error("HPA list should not be empty.")
	}
}

func TestHPAGet(t *testing.T) {
	fakeHPAClient := createFakeHPAClient()
	if _, err := fakeHPAClient.Get("fakehpa"); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if hpa, err := fakeHPAClient.Get("nonexistent"); err == nil {
		t.Errorf("Unexpected hpa: %v", hpa)
	}
}

func TestHPAExist(t *testing.T) {
	fakeHPAClient := createFakeHPAClient()
	if _, err := fakeHPAClient.Get("fakehpa"); err != nil {
		t.Errorf("Failed with error: %s", err.Error())
	}
}

func createFakeHPAClient() HorizontalPodAutoscaler {
	return HorizontalPodAutoscaler{
		Interface: createFakeHPAClientset(),
		Namespace: "sample",
	}
}

func createFakeHPAClientset() *fake.Clientset {
	var min, target int32 = 1, 50
	return fake.NewSimpleClientset(
		&autoscale_v1.HorizontalPodAutoscaler{
			ObjectMeta: v1.ObjectMeta{
				Name:      "fakehpa",
				Namespace: "sample",
				Labels: map[string]string{
					"creator":     "pipeline",
					"name":        "fakehpa",
					"application": "myapp",
					"version":     "0.1",
				},
			},
			Spec: autoscale_v1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscale_v1.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       "fakedeployment",
					APIVersion: "v1beta1",
				},
				MinReplicas:                    &min,
				MaxReplicas:                    int32(3),
				TargetCPUUtilizationPercentage: &target,
			},
		},
	)
}
