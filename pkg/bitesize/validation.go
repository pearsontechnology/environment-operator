package bitesize

import (
	"fmt"
	"reflect"

	"github.com/pearsontechnology/environment-operator/pkg/config"
)

func validVolumeModes(v interface{}, param string) error {
	validNames := map[string]bool{"ReadWriteOnce": true, "ReadOnlyMany": true, "ReadWriteMany": true}
	st := reflect.ValueOf(v)

	if st.Kind() != reflect.String {
		return fmt.Errorf(
			"Invalid volume mode: %v. Valid modes: %s",
			st,
			"ReadWriteOnce,ReadOnlyMany,ReadWriteMany",
		)
	}

	if validNames[st.String()] == false {
		return fmt.Errorf("Invalid volume mode: %v", st)
	}
	return nil
}

func validHPA(hpa interface{}, param string) error {
	val := reflect.ValueOf(hpa)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).Int()
		fieldName := val.Type().Field(i).Name

		switch fieldName {

		case "MinReplicas", "MaxReplicas":
			if fieldValue != 0 {
				if fieldValue > int64(config.Env.HPAMaxReplicas) {
					return fmt.Errorf("hpa %+v number of replicas invalid; values greater than %v not allowed", hpa, config.Env.HPAMaxReplicas)
				}
			}

		case "TargetCPUUtilizationPercentage":
			if fieldValue != 0 {
				if fieldValue < 75 {
					return fmt.Errorf("hpa %+v CPU Utilization invalid; thresholds lower than 75%% not allowed", hpa)
				}
			}
		}
	}

	return nil
}
