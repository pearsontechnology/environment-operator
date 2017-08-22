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

func TestValidHPA(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 51, TargetCPUUtilizationPercentage: 75},
			fmt.Errorf("hpa %+v number of replicas invalid; values greater than %v not allowed", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 51, TargetCPUUtilizationPercentage: 75}, config.Env.HPAMaxReplicas),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, TargetCPUUtilizationPercentage: 74},
			fmt.Errorf("hpa %+v CPU Utilization invalid; thresholds lower than 75%% not allowed", HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, TargetCPUUtilizationPercentage: 74}),
		},
		{
			HorizontalPodAutoscaler{MinReplicas: 1, MaxReplicas: 2, TargetCPUUtilizationPercentage: 75},
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
			ContainerRequests{CPU: "1000m"},
			fmt.Errorf("requests %+v invalid CPU quantity; values greater than %vm not allowed", ContainerRequests{CPU: "1000m"}, config.Env.ReqMaxCPU),
		},
		{
			ContainerRequests{CPU: "500m"},
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
