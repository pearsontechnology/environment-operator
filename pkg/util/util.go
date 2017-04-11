package util

import (
	"fmt"
	"os"
)

// func trimBlueGreenFromName(orig string) string {
// 	return strings.TrimSuffix(strings.TrimSuffix(orig, "-blue"), "-green")
// }
//
// func trimBlueGreenFromHost(orig string) string {
// 	split := strings.Split(orig, ".")
// 	split[0] = trimBlueGreenFromName(split[0])
// 	return strings.Join(split, ".")
// }
// func collectHealthCheck(probe *v1.Probe) *bitesize.HealthCheck {
// 	return &bitesize.HealthCheck{}
// }

// Registry returns docker registry setting
func Registry() string {
	return os.Getenv("DOCKER_REGISTRY")
}

// Project returns project's name. TODO: Should be loaded from namespace labels..
func Project() string {
	return os.Getenv("PROJECT")
}

// ApplicationImage returns full image name given bitesize.Service object
// func ApplicationImage(svc *bitesize.Service) (string, error) {
// 	return Image(svc.Application, svc.Version), nil
// }

// Image returns full  app image given app and version
func Image(app, version string) string {
	return fmt.Sprintf(
		"%s/%s/%s:%s",
		Registry(), Project(), app, version,
	)
}

func EqualArrays(a, b []int) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
