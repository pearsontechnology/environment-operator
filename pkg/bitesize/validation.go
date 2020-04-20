package bitesize

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	validator "gopkg.in/validator.v2"
)

func addCustomValidators() {
	validator.SetValidationFunc("volume_modes", validVolumeModes)
	validator.SetValidationFunc("volume_types", validVolumeTypes)
	validator.SetValidationFunc("volume_provisioning", validVolumeProvisioning)
	validator.SetValidationFunc("hpa", validHPA)
	validator.SetValidationFunc("requests", validRequests)
	validator.SetValidationFunc("limits", validLimits)
	validator.SetValidationFunc("external_url", validExternalURL)
}

func validVolumeModes(v interface{}, param string) error {
	validNames := map[string]bool{"ReadWriteOnce": true, "ReadOnlyMany": true, "ReadWriteMany": true}
	str, ok := v.(string)

	if !ok {
		return fmt.Errorf(
			"Invalid volume mode: %v. Valid modes: %s",
			v,
			"ReadWriteOnce,ReadOnlyMany,ReadWriteMany",
		)
	}

	if _, ok := validNames[str]; !ok {
		return fmt.Errorf("Invalid volume mode: %s", str)
	}
	return nil
}

func validVolumeTypes(v interface{}, param string) error {
	validTypes := map[string]bool{"efs": true, "EBS": true}
	str, ok := v.(string)

	if !ok { // not a string
		return fmt.Errorf(
			"Invalid volume type: %v. Valid types: %s",
			v,
			"EBS,EFS",
		)
	}

	if _, ok := validTypes[str]; !ok { // invalid
		return fmt.Errorf(
			"Invalid volume type: %v. Valid types: %s",
			v,
			"EBS,EFS",
		)
	}
	return nil
}

func validVolumeProvisioning(v interface{}, param string) error {
	validProvisioningTypes := map[string]bool{"dynamic": true, "manual": true}
	st := reflect.ValueOf(v)

	if st.Kind() != reflect.String {
		return fmt.Errorf(
			"Invalid provisioning type: %v. Valid types: %s",
			st,
			"dynamic,manual",
		)
	}

	if validProvisioningTypes[st.String()] == false {
		return fmt.Errorf("Invalid provisioning type: %v", st)
	}
	return nil
}

func validHPA(hpa interface{}, param string) error {
	val := reflect.ValueOf(hpa)

	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name

		switch fieldName {

		case "MinReplicas", "MaxReplicas":
			fieldValue := val.Field(i).Int()
			if fieldValue != 0 && fieldValue > int64(config.Env.HPAMaxReplicas) {
				return fmt.Errorf("hpa %+v number of replicas invalid; values greater than %v not allowed", hpa, config.Env.HPAMaxReplicas)
			}

		case "Metric":
			fieldValue := val.Field(i).Interface()

			metric, ok := fieldValue.(Metric)
			if ok {
				if metric.TargetAverageUtilization != 0 && metric.TargetAverageUtilization < 75 {
					return fmt.Errorf("hpa %+v CPU Utilization invalid; thresholds lower than 75%% not allowed", hpa)
				}
				if metric.TargetAverageUtilization != 0 && metric.TargetAverageUtilization > 100 {
					return fmt.Errorf("hpa %+v memory Utilization invalid; thresholds greater than 100%% not allowed", hpa)
				}
				if metric.Name != "cpu" && metric.Name != "memory" {
					if metric.TargetAverageUtilization != 0 {
						return fmt.Errorf("hpa %+v target Average Utilization does not exist for custom metrics", hpa)
					}
				}
			}
		}
	}

	return nil
}

func validRequests(req interface{}, param string) error {
	//TODO: Add other supported unit types
	validUnits := map[string]bool{
		"Mi": true,
		"m":  true,
	}
	val := reflect.ValueOf(req)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).String()
		fieldName := val.Type().Field(i).Name

		switch fieldName {

		case "CPU":
			if fieldValue != "" {
				unit := string(fieldValue[len(fieldValue)-1:])
				if !validUnits[unit] {
					log.Debugf("requests %+v invalid CPU units; "+`"m"`+" suffix not specified", req)
					return fmt.Errorf("requests %+v invalid CPU units; "+`"m"`+" suffix not specified", req)
				}
				if quantity, _ := strconv.Atoi(fieldValue[0 : len(fieldValue)-1]); quantity > config.Env.LimitMaxCPU {
					log.Debugf("requests %+v invalid CPU quantity; values greater than maximum CPU limit %vm not allowed", req, config.Env.LimitMaxCPU)
					return fmt.Errorf("requests %+v invalid CPU quantity; values greater than maximum CPU limit %vm not allowed", req, config.Env.LimitMaxCPU)
				}
			}

		case "Memory":
			if fieldValue != "" {
				unit := string(fieldValue[len(fieldValue)-2:])
				if !validUnits[unit] {
					log.Debugf("requests %+v invalid memory units; "+`"Mi"`+" suffix not specified", req)
					return fmt.Errorf("requests %+v invalid memory units; "+`"Mi"`+" suffix not specified", req)
				}
				if quantity, _ := strconv.Atoi(fieldValue[0 : len(fieldValue)-2]); quantity > config.Env.LimitMaxMemory {
					log.Debugf("requests %+v invalid memory quantity; values greater than maximum memory limit %vMi not allowed", req, config.Env.LimitMaxMemory)
					return fmt.Errorf("requests %+v invalid memory quantity; values greater than maximum memory limit %vMi not allowed", req, config.Env.LimitMaxMemory)
				}
			}

		}

	}

	return nil
}

func validLimits(req interface{}, param string) error {
	//TODO: Add other supported unit types
	validUnits := map[string]bool{
		"Mi": true,
		"m":  true,
	}
	val := reflect.ValueOf(req)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).String()
		fieldName := val.Type().Field(i).Name

		switch fieldName {

		case "CPU":
			if fieldValue != "" {
				unit := string(fieldValue[len(fieldValue)-1:])
				if !validUnits[unit] {
					log.Debugf("limits %+v invalid CPU units; "+`"m"`+" suffix not specified", req)
					return fmt.Errorf("limits %+v invalid CPU units; "+`"m"`+" suffix not specified", req)
				}
				if quantity, _ := strconv.Atoi(fieldValue[0 : len(fieldValue)-1]); quantity > config.Env.LimitMaxCPU {
					log.Debugf("limits %+v invalid CPU quantity; values greater than %vm not allowed", req, config.Env.LimitMaxCPU)
					return fmt.Errorf("limits %+v invalid CPU quantity; values greater than %vm not allowed", req, config.Env.LimitMaxCPU)
				}
			}

		case "Memory":
			if fieldValue != "" {
				unit := string(fieldValue[len(fieldValue)-2:])
				if !validUnits[unit] {
					log.Debugf("limits %+v invalid Memory units; "+`"Mi"`+" suffix not specified", req)
					return fmt.Errorf("limits %+v invalid Memory units; "+`"Mi"`+" suffix not specified", req)
				}
				if quantity, _ := strconv.Atoi(fieldValue[0 : len(fieldValue)-2]); quantity > config.Env.LimitMaxMemory {
					log.Debugf("limits %+v invalid Memory quantity; values greater than %vMi not allowed", req, config.Env.LimitMaxMemory)
					return fmt.Errorf("limits %+v invalid Memory quantity; values greater than %vMi not allowed", req, config.Env.LimitMaxMemory)
				}
			}

		}
	}

	return nil
}

func validExternalURL(urls interface{}, param string) error {
	urlSlice := reflect.ValueOf(urls)
	for i := 0; i < urlSlice.Len(); i++ {
		url := urlSlice.Index(i).String()
		ok, err := regexp.MatchString("^([a-zA-Z0-9\\.\\-]+)*$", url)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("external_url %v is invalid", url)
		}
	}
	return nil
}
