package bitesize

import (
	"fmt"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/config"
)

func TestValidationVolumeNames(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			"redbluegreen",
			fmt.Errorf("Invalid volume mode: redbluegreen"),
		},
		{
			1,
			fmt.Errorf("Invalid volume mode: 1. Valid modes: ReadWriteOnce,ReadOnlyMany,ReadWriteMany"),
		},
		{
			"ReadWriteOnce",
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validVolumeModes(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

}

func TestValidVolumeProvisioning(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			"redbluegreen",
			fmt.Errorf("Invalid provisioning type: redbluegreen"),
		},
		{
			1,
			fmt.Errorf("Invalid provisioning type: 1. Valid types: dynamic,manual"),
		},
		{
			"manual",
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validVolumeProvisioning(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

}

func TestValidHPA(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 51, Metric: Metric{Name: "cpu", TargetAverageUtilization: 75}},
			fmt.Errorf("hpa %+v number of replicas invalid; values greater than %v not allowed", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 51, Metric: Metric{Name: "cpu", TargetAverageUtilization: 75}}, config.Env.HPAMaxReplicas),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "cpu", TargetAverageUtilization: 74}},
			fmt.Errorf("hpa %+v CPU Utilization invalid; thresholds lower than 75%% not allowed", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "cpu", TargetAverageUtilization: 74}}),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "memory", TargetAverageUtilization: 120}},
			fmt.Errorf("hpa %+v memory Utilization invalid; thresholds greater than 100%% not allowed", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "memory", TargetAverageUtilization: 120}}),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "custom_metric", TargetAverageUtilization: 80}},
			fmt.Errorf("hpa %+v target Average Utilization does not exist for custom metrics", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "custom_metric", TargetAverageUtilization: 80}}),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, Metric: Metric{Name: "cpu", TargetAverageUtilization: 75}},
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validHPA(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("HPA validation error: %v", err)
			}
		}
	}

}

func TestValidRequests(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			ContainerRequests{CPU: "100"},
			fmt.Errorf("requests %+v invalid CPU units; "+`"m"`+" suffix not specified", ContainerRequests{CPU: "100"}),
		},
		{
			ContainerRequests{CPU: "5000m"},
			fmt.Errorf("requests %+v invalid CPU quantity; values greater than maximum CPU limit %vm not allowed", ContainerRequests{CPU: "5000m"}, config.Env.LimitMaxCPU),
		},
		{
			ContainerRequests{CPU: "500m"},
			nil,
		},
		{
			ContainerRequests{Memory: "100"},
			fmt.Errorf("requests %+v invalid memory units; "+`"Mi"`+" suffix not specified", ContainerRequests{Memory: "100"}),
		},
		{
			ContainerRequests{Memory: "9000Mi"},
			fmt.Errorf("requests %+v invalid memory quantity; values greater than maximum memory limit %vMi not allowed", ContainerRequests{Memory: "9000Mi"}, config.Env.LimitMaxMemory),
		},
		{
			ContainerRequests{Memory: "500Mi"},
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validRequests(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("Requests validation error: %v", err)
			}
		}
	}

}

func TestValidLimits(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			ContainerLimits{CPU: "100"},
			fmt.Errorf("limits %+v invalid CPU units; "+`"m"`+" suffix not specified", ContainerLimits{CPU: "100"}),
		},
		{
			ContainerLimits{CPU: "5000m"},
			fmt.Errorf("limits %+v invalid CPU quantity; values greater than %vm not allowed", ContainerLimits{CPU: "5000m"}, config.Env.LimitMaxCPU),
		},
		{
			ContainerLimits{CPU: "500m"},
			nil,
		},
		{
			ContainerLimits{Memory: "100"},
			fmt.Errorf("limits %+v invalid Memory units; "+`"Mi"`+" suffix not specified", ContainerLimits{Memory: "100"}),
		},
		{
			ContainerLimits{Memory: "9000Mi"},
			fmt.Errorf("limits %+v invalid Memory quantity; values greater than %vMi not allowed", ContainerLimits{Memory: "9000Mi"}, config.Env.LimitMaxMemory),
		},
		{
			ContainerLimits{Memory: "500Mi"},
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validLimits(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("Limits validation error: %v", err)
			}
		}
	}

}

func TestValidExternalURL(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			[]string{"~www.example.com"},
			fmt.Errorf("external_url %v is invalid", "~www.example.com"),
		},
		{
			[]string{"www.test.com"},
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validExternalURL(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("external_url validation error: %v", err)
			}
		}
	}

}
