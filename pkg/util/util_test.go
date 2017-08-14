package util

import (
	"os"
	"testing"
)

func TestEnvironmentVariableUtility(t *testing.T) {
	os.Setenv("DOCKER_PULL_SECRETS", "MySecret")

	if RegistrySecrets() != "MySecret" {
		t.Errorf("Unexpected Variable retrieved for DOCKER_PULL_SECRETS")
	}
}
