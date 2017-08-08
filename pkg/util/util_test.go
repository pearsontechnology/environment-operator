package util

import (
	"github.com/pearsontechnology/environment-operator/pkg/util"
	"os"
	"testing"
)

func TestEnvironmentVariableUtility(t *testing.T) {
	os.Setenv("DOCKER_PULL_SECRETS", "MySecret")

	if util.RegistrySecrets() != "MySecret" {
		t.Errorf("Unexpected Variable retrieved for DOCKER_PULL_SECRETS")
	}
}
