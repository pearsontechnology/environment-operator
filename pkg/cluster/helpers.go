package cluster

import (
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	v1beta2_apps "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	v1beta1_ext "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func envVars(deployment v1beta1_ext.Deployment) []bitesize.EnvVar {
	var retval []bitesize.EnvVar
	for _, e := range deployment.Spec.Template.Spec.Containers[0].Env {
		var v bitesize.EnvVar
		// Reserved vars
		if isReservedEnvVar(e) {
			continue
		}

		if e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil {
			v = bitesize.EnvVar{
				Value:  e.ValueFrom.SecretKeyRef.Key,
				Secret: e.Name,
			}
		} else {
			v = bitesize.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			}
		}
		retval = append(retval, v)
	}
	return retval
}

func isReservedEnvVar(e v1.EnvVar) bool {
	reserved := []string{"POD_DEPLOYMENT_COLOUR"}
	for _, i := range reserved {
		if e.Name == i {
			return true
		}
	}
	return false
}

func envVarsStatefulset(statefulset v1beta2_apps.StatefulSet) []bitesize.EnvVar {
	var retval []bitesize.EnvVar
	for _, e := range statefulset.Spec.Template.Spec.Containers[0].Env {
		var v bitesize.EnvVar

		if isReservedEnvVar(e) {
			continue
		}

		if e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil {
			v = bitesize.EnvVar{
				Value:  e.ValueFrom.SecretKeyRef.Key,
				Secret: e.Name,
			}
		} else {
			v = bitesize.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			}
		}
		retval = append(retval, v)
	}
	return retval
}
func healthCheck(deployment v1beta1_ext.Deployment) *bitesize.HealthCheck {
	var retval *bitesize.HealthCheck

	probe := deployment.Spec.Template.Spec.Containers[0].LivenessProbe
	if probe != nil && probe.Exec != nil {

		retval = &bitesize.HealthCheck{
			Command:      probe.Exec.Command,
			InitialDelay: int(probe.InitialDelaySeconds),
			Timeout:      int(probe.TimeoutSeconds),
		}
	}
	return retval
}
func healthCheckStatefulset(statefulset v1beta2_apps.StatefulSet) *bitesize.HealthCheck {
	var retval *bitesize.HealthCheck

	probe := statefulset.Spec.Template.Spec.Containers[0].LivenessProbe
	if probe != nil && probe.Exec != nil {

		retval = &bitesize.HealthCheck{
			Command:      probe.Exec.Command,
			InitialDelay: int(probe.InitialDelaySeconds),
			Timeout:      int(probe.TimeoutSeconds),
		}
	}
	return retval
}
func getLabel(metadata metav1.ObjectMeta, label string) string {
	labels := metadata.GetLabels()
	return labels[label]
}

func getAnnotation(metadata metav1.ObjectMeta, annotation string) string {
	annotations := metadata.GetAnnotations()
	return annotations[annotation]
}

func getAccessModesAsString(modes []v1.PersistentVolumeAccessMode) string {

	modesStr := []string{}
	if containsAccessMode(modes, v1.ReadWriteOnce) {
		modesStr = append(modesStr, "ReadWriteOnce")
	}
	if containsAccessMode(modes, v1.ReadOnlyMany) {
		modesStr = append(modesStr, "ReadOnlyMany")
	}
	if containsAccessMode(modes, v1.ReadWriteMany) {
		modesStr = append(modesStr, "ReadWriteMany")
	}
	return strings.Join(modesStr, ",")
}

func containsAccessMode(modes []v1.PersistentVolumeAccessMode, mode v1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}
