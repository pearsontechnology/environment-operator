package cluster

import (
	"strings"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func envVars(deployment apps_v1.Deployment) []bitesize.EnvVar {
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

func healthCheck(deployment apps_v1.Deployment) *bitesize.HealthCheck {
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

func convertProbeType(probe *v1.Probe) *bitesize.Probe {
	var retval *bitesize.Probe

	if probe != nil {
		retval = &bitesize.Probe{
			FailureThreshold:    probe.FailureThreshold,
			InitialDelaySeconds: probe.InitialDelaySeconds,
			SuccessThreshold:    probe.SuccessThreshold,
			TimeoutSeconds:      probe.TimeoutSeconds,
			PeriodSeconds:       probe.PeriodSeconds,
		}

		if probe.HTTPGet != nil {
			httpGet := &bitesize.HTTPGetAction{}

			for _, v := range probe.HTTPGet.HTTPHeaders {
				httpGet.HTTPHeaders = append(httpGet.HTTPHeaders, bitesize.HTTPHeader{
					Name:  v.Name,
					Value: v.Value,
				})
			}

			httpGet.Host = probe.HTTPGet.Host
			httpGet.Path = probe.HTTPGet.Path
			httpGet.Port = probe.HTTPGet.Port.IntVal
			httpGet.Scheme = probe.HTTPGet.Scheme

			retval.Handler = bitesize.Handler{
				HTTPGet: httpGet,
			}
		}

		if probe.Exec != nil {
			exec := &bitesize.ExecAction{}

			exec.Command = probe.Exec.Command

			retval.Handler = bitesize.Handler{
				Exec: exec,
			}
		}

		if probe.TCPSocket != nil {
			socket := &bitesize.TCPSocketAction{}

			socket.Host = probe.TCPSocket.Host
			socket.Port = probe.TCPSocket.Port.IntVal

			retval.Handler = bitesize.Handler{
				TCPSocket: socket,
			}
		}
	}
	return retval
}

func livenessProbe(deployment apps_v1.Deployment) *bitesize.Probe {
	probe := deployment.Spec.Template.Spec.Containers[0].LivenessProbe

	return convertProbeType(probe)
}

func readinessProbe(deployment apps_v1.Deployment) *bitesize.Probe {
	probe := deployment.Spec.Template.Spec.Containers[0].ReadinessProbe

	return convertProbeType(probe)
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

	var modesStr []string
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

func volumes(deployment apps_v1.Deployment) []bitesize.Volume {
	//TODO: implement other volume types
	var volumes []bitesize.Volume

	volumeMounts := deployment.Spec.Template.Spec.Containers[0].VolumeMounts
	// add ConfigMap volumes to diff
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		// if ConfigMap volume
		if v.VolumeSource.ConfigMap != nil {
			vol := bitesize.Volume{
				Name:  v.Name,
				Type:  bitesize.TypeConfigMap,
				Modes: "ReadWriteOnce",
			}
			// find the mount path for the volume
			for _, mount := range volumeMounts {
				if mount.Name == v.Name {
					vol.Path = mount.MountPath
				}
			}
			// generate items if any
			for _, it := range v.ConfigMap.Items {
				vol.Items = append(vol.Items, bitesize.KeyToPath{
					Key:  it.Key,
					Path: it.Path,
					Mode: it.Mode,
				})
			}
			volumes = append(volumes, vol)
		}
	}
	return volumes
}
