package util

import (
	"fmt"
	"os"
)

// Registry returns docker registry setting
func Registry() string {
	return os.Getenv("DOCKER_REGISTRY")
}

func RegistrySecrets() string {
	return os.Getenv("DOCKER_PULL_SECRETS")
}

// Project returns project's name. TODO: Should be loaded from namespace labels..
func Project() string {
	return os.Getenv("PROJECT")
}

// Image returns full  app image given app and version
func Image(app, version string) string {
	if Registry() == "" {
		return fmt.Sprintf("%s/%s:%s", Project(), app, version)
	}

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
