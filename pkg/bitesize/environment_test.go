package bitesize

import (
	"reflect"
	"sort"
	"testing"
)

func TestExistingEnvironment(t *testing.T) {

	e, err := LoadEnvironment("../../test/assets/environments.bitesize", "environment2")

	if err != nil {
		t.Errorf("Unexpected error loading environment: %s", err.Error())
	}

	if len(e.Services) != 6 {
		t.Errorf("Unexpected count of services. Expected 6, got: %d", len(e.Services))
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
		t.Errorf("Environment sort invalid, got %q", e)
	}
}
