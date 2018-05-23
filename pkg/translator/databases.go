package translator

import (
	"math/rand"

	"k8s.io/client-go/pkg/api/v1"
)

// In this file we define database specific containers, volumes and annotations used to generate statefulset objects

// Couchbase

// CbContainers generates a slice containing the couchbase containers required in a couchbase statefulset
func (w *KubeMapper) CbContainers() ([]v1.Container, error) {
	mounts, err := w.volumeMounts()
	if err != nil {
		return nil, err
	}
	res, err := w.resources()
	if err != nil {
		return nil, err
	}

	cb := []v1.Container{
		{
			Name:            "couchbase",
			Image:           "couchbase/server:enterprise-" + w.BiteService.Version,
			ImagePullPolicy: v1.PullIfNotPresent,
			Ports: []v1.ContainerPort{
				{ContainerPort: int32(8091), Name: "cb-admin"},
				{ContainerPort: int32(8092), Name: "cb-views"},
				{ContainerPort: int32(8093), Name: "cb-queries"},
				{ContainerPort: int32(8094), Name: "cb-search"},
				{ContainerPort: int32(9100), Name: "cb-int-ind-ad"},
				{ContainerPort: int32(9101), Name: "cb-int-ind-sc"},
				{ContainerPort: int32(9102), Name: "cb-int-ind-ht"},
				{ContainerPort: int32(9103), Name: "cb-int-ind-in"},
				{ContainerPort: int32(9104), Name: "cb-int-ind-ca"},
				{ContainerPort: int32(9105), Name: "cb-int-ind-ma"},
				{ContainerPort: int32(9998), Name: "cb-int-rest"},
				{ContainerPort: int32(9999), Name: "cb-int-gsi"},
				{ContainerPort: int32(11207), Name: "cb-memc-ssl"},
				{ContainerPort: int32(11209), Name: "cb-int-bu"},
				{ContainerPort: int32(11210), Name: "cb-moxi"},
				{ContainerPort: int32(11211), Name: "cb-memc"},
				{ContainerPort: int32(11214), Name: "cb-ssl-xdr1"},
				{ContainerPort: int32(11215), Name: "cb-ssl-xdr2"},
				{ContainerPort: int32(18091), Name: "cb-admin-ssl"},
				{ContainerPort: int32(18092), Name: "cb-views-ssl"},
				{ContainerPort: int32(18093), Name: "cb-queries-ssl"},
				{ContainerPort: int32(4369), Name: "empd"},
			},
			VolumeMounts: mounts,
			Resources:    res,
		},
		{
			Name:            "couchbase-sidecar",
			Image:           "pearsontechnology/couchbase-sidecar:latest",
			ImagePullPolicy: v1.PullAlways,
			Lifecycle: &v1.Lifecycle{
				PreStop: &v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"couchbase-sidecar", "-remove-node", "true"},
					},
				},
			},
			Env: []v1.EnvVar{
				{Name: "SIDECAR_SERVICE", Value: w.BiteService.Name},
				{Name: "SIDECAR_MASTERNODE", Value: w.BiteService.Name + "-0"},
				{Name: "SIDECAR_HOST", ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				},
				{Name: "SIDECAR_NAMESPACE", ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{Name: "SIDECAR_ADMINPASSWORD", ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: w.BiteService.Name,
						},
						Key: "admin_password",
					},
				},
				},
				{Name: "SIDECAR_CLIENTPASSWORD", ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: w.BiteService.Name,
						},
						Key: "client_password",
					},
				},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					MountPath: "/opt/couchbase/var/lib/couchbase/backups",
					Name:      "backups",
				},
			},
		},
	}
	return cb, err
}

// CbSecret generates a secret containing admin and client credentials for couchbase
func (w *KubeMapper) CbSecret() v1.Secret {
	s := map[string]string{
		"admin_password":  randomPassword(12),
		"client_password": randomPassword(12),
	}

	return v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name,
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":    "pipeline",
				"deployment": w.BiteService.Name,
			},
		},
		StringData: s,
	}
}

// CbUIService generates a service object for couchbase UI access
func (w *KubeMapper) CbUIService() *v1.Service {
	// This is used as a workaround since the UI requires persistent sessions
	// and our ingress controller does not support this
	// so we just point to the first pod in the statefulset for UI access
	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      w.BiteService.Name + "-ui",
			Namespace: w.Namespace,
			Labels: map[string]string{
				"creator":     "pipeline",
				"name":        w.BiteService.Name + "-ui",
				"application": w.BiteService.Application,
				// otherwise reaper will delete it :(
				"delete-protected": "yes",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceType("ExternalName"),
			ExternalName: "cb-0" + "." + w.BiteService.Name + "." + w.Namespace + ".svc.cluster.local",
		},
	}
}

func randomPassword(length int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
