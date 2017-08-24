package bitesize

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pearsontechnology/environment-operator/pkg/config"

	validator "gopkg.in/validator.v2"
)

func addCustomValidators() {
	validator.SetValidationFunc("volume_modes", validVolumeModes)
	validator.SetValidationFunc("hpa", validHPA)
	validator.SetValidationFunc("requests", validRequests)
}

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
			if fieldValue != 0 && fieldValue > int64(config.Env.HPAMaxReplicas) {
				return fmt.Errorf("hpa %+v number of replicas invalid; values greater than %v not allowed", hpa, config.Env.HPAMaxReplicas)
			}

		case "TargetCPUUtilizationPercentage":
			if fieldValue != 0 && fieldValue < 75 {
				return fmt.Errorf("hpa %+v CPU Utilization invalid; thresholds lower than 75%% not allowed", hpa)
			}
		}
	}

	return nil
}

func validRequests(req interface{}, param string) error {
	val := reflect.ValueOf(req)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).String()
		fieldName := val.Type().Field(i).Name

		switch fieldName {

		case "CPU":
			if fieldValue != "" {
				if unit := string(fieldValue[len(fieldValue)-1:]); unit != "m" {
					return fmt.Errorf("requests %+v invalid CPU units; "+`"m"`+" suffix not specified", req)
				}
				if quantity, _ := strconv.Atoi(fieldValue[0 : len(fieldValue)-1]); quantity > config.Env.ReqMaxCPU {
					return fmt.Errorf("requests %+v invalid CPU quantity; values greater than %vm not allowed", req, config.Env.ReqMaxCPU)
				}
			}

		}
	}

	return nil
}
