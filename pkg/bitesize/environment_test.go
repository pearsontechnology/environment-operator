package bitesize

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestExistingEnvironment(t *testing.T) {

	e, err := LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Errorf("TestExistingEnvironment: Unexpected error loading environment: %s", err.Error())
	}

	if len(e.Services) != 7 {
		t.Errorf("Unexpected count of services. Expected 7, got: %d", len(e.Services))
	}

}

func TestNoneExistingEnvironment(t *testing.T) {
	e, err := LoadEnvironment("../../test/assets/environments.bitesize", "non-existant")
	if e != nil {
		t.Errorf("Expected environment to be nil, got %v", e)
	}
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestNonExistingEnvironmentFile(t *testing.T) {
	e, err := LoadEnvironment("/nonexisting", "blah")
	if e != nil {
		t.Errorf("Expected environment to be nil, got: %v", e)
	}
	if err.Error() != "open /nonexisting: no such file or directory" {
		t.Errorf("Expected error, got %s", err.Error())
	}
}

func TestEnvironmentSortInterface(t *testing.T) {
	var e = Environments{
		{Name: "b"},
		{Name: "a"},
		{Name: "c"},
	}

	var expected = Environments{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}

	sort.Sort(e)
	if !reflect.DeepEqual(e, expected) {
		t.Errorf("Environment sort invalid, got %v", e)
	}
}

func TestEnvironmentBlueGreenLoading(t *testing.T) {
	e, err := LoadEnvironment("../../test/assets/environments.bitesize", "environment11")
	if err != nil {
		t.Errorf("Unexpected error when loading environment: %s", err.Error())
	}

	if len(e.Services) != 3 {
		t.Errorf("Unexpected environment count: expected 3, got %d", len(e.Services))
	}
}

func TestEnvironmentImportConfigMap(t *testing.T) {
	config.Env.UseAuth = false

	e, err := LoadEnvironment("../../test/assets/environments3.bitesize", "environment1")

	// fails on travis because .git in travis
	if err != nil {
		t.Errorf("TestEnvironmentImportConfigMap: Unexpected error loading environment: %s", err.Error())
	}

	if len(e.Gists) != 3 {
		t.Errorf("Unexpected count of import. Expected 3, got: %d", len(e.Gists))
	}
}

func TestFindVolByName(t *testing.T) {
	volumes := []Volume{
		{
			Name: "zzz",
			Type: "secret",
			Path: "path-zzz",
		},
		{
			Name:  "aaa",
			Type:  "efs",
			Path:  "path-aaa",
			Size:  "10G",
			Modes: "ReadWriteMany",
		},
		{
			Name:  "mmm",
			Size:  "10G",
			Path:  "path-mmm",
			Modes: "ReadWriteOnce",
		},
	}

	vol, err := findVolByName(volumes, "aaa")
	if err != nil {
		t.Errorf("Could not find volume name: %s", "aaa")
	}
	assert.Equal(t, vol, volumes[1], "The returned volume should be the same")
}

func TestSortVolumesByVolName(t *testing.T) {
	volumesInput := []Volume{
		{
			Name: "zzz",
			Type: "secret",
			Path: "path-zzz",
		},
		{
			Name:  "aaa",
			Type:  "efs",
			Path:  "path-aaa",
			Size:  "10G",
			Modes: "ReadWriteMany",
		},
		{
			Name:  "mmm",
			Size:  "10G",
			Path:  "path-mmm",
			Modes: "ReadWriteOnce",
		},
	}

	volumesExpected := []Volume{
		{
			Name:  "aaa",
			Type:  "efs",
			Path:  "path-aaa",
			Size:  "10G",
			Modes: "ReadWriteMany",
		},
		{
			Name:  "mmm",
			Size:  "10G",
			Path:  "path-mmm",
			Modes: "ReadWriteOnce",
		},
		{
			Name: "zzz",
			Type: "secret",
			Path: "path-zzz",
		},
	}
	volumesSorted, err := SortVolumesByVolName(volumesInput)
	if err != nil {
		t.Errorf("Could not sort volumes: %s", err)
	}
	assert.Equal(t, volumesExpected, volumesSorted, "The volumes should be sorted")
}
