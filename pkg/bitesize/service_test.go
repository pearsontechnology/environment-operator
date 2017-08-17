package bitesize

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/util"

	yaml "gopkg.in/yaml.v2"
)

// Tests to see that YAML documents unmarshal correctly

func TestUnmarshalPorts(t *testing.T) {
	t.Run("ports string parsed correctly", testPortsString)
	t.Run("ports preferred over port", testPortsOverPort)
	t.Run("ports with invalid value", testPortsWithInvalidValue)
	t.Run("empty ports return default", testPortsEmpty)
}

func TestFindByName(t *testing.T) {
	t.Run("find existing service", testFindByNameExist)
	t.Run("find non-existing service", testFindByNameNotFound)
}

func testPortsString(t *testing.T) {
	svc := &Service{}
	str := `
  name: something
  ports: 81,88,89
  `
	if err := yaml.Unmarshal([]byte(str), svc); err != nil {
		t.Errorf("could not unmarshal yaml: %s", err.Error())
	}

	if !util.EqualArrays(svc.Ports, []int{81, 88, 89}) {
		t.Errorf("Ports not equal. Expected: [81 88 89], got: %v", svc.Ports)
	}

}

func testPortsOverPort(t *testing.T) {
	svc := &Service{}
	str := `
  name: something
  port: 80
  ports: 81,82
  `

	if err := yaml.Unmarshal([]byte(str), svc); err != nil {
		t.Errorf("could not unmarshal yaml: %s", err.Error())
	}

	if len(svc.Ports) != 2 {
		t.Errorf("Unexpected ports: %v", svc.Ports)
	}

}

func testPortsWithInvalidValue(t *testing.T) {
	svc := &Service{}
	str := `
  name: something
  ports: 81,invalid,82
  `
	if err := yaml.Unmarshal([]byte(str), svc); err != nil {
		t.Errorf("could not unmarshal yaml: %s", err.Error())
	}

	if !util.EqualArrays(svc.Ports, []int{81, 82}) {
		t.Errorf("Unexpected ports: %v", svc.Ports)
	}
}

func testPortsEmpty(t *testing.T) {
	svc := &Service{}
	str := `
  name: something
  `
	if err := yaml.Unmarshal([]byte(str), svc); err != nil {
		t.Errorf("could not unmarshal yaml: %s", err.Error())
	}

	if !util.EqualArrays(svc.Ports, []int{80}) {
		t.Errorf("Unexpected ports: %v", svc.Ports)
	}
}
func TestPod(t *testing.T) {
	t.Run("test Pods exist in Service", testPodsEqual)
}
func testPodsEqual(t *testing.T) {
	e1 := Pod{Name: "PodA", Phase: "Running", StartTime: "StartTime", Message: "Message", Logs: "Log Message"}
	e2 := Pod{Name: "PodA", Phase: "Running", StartTime: "StartTime", Message: "Message", Logs: "Log Message"}

	if e1 != e2 {
		t.Errorf("Expected %+v to be equal to %+v, got false", e1, e2)
	}
}

func TestEnvVars(t *testing.T) {
	t.Run("test equal env vars", testEnvVarsEqual)
	t.Run("test diff env vars", testEnvVarsDiff)
}

func testEnvVarsEqual(t *testing.T) {
	e1 := EnvVar{Name: "a", Value: "1"}
	e2 := EnvVar{Name: "a", Value: "1"}

	if e1 != e2 {
		t.Errorf("Expected %+v to be equal to %+v, got false", e1, e2)
	}

	e1 = EnvVar{Secret: "secret", Value: "zzz"}
	e2 = EnvVar{Secret: "secret", Value: "zzz"}

	if e1 != e2 {
		t.Errorf("Expected %+v to be equal to %+v, got false", e1, e2)
	}
}

func testEnvVarsDiff(t *testing.T) {
	e1 := EnvVar{Name: "a", Value: "1"}
	e2 := EnvVar{Name: "a", Value: "2"}
	if e1 == e2 {
		t.Errorf("Expected %+v to be not equal to %+v, got true", e1, e2)
	}

	e1 = EnvVar{Secret: "secret", Value: "zzz"}
	e2 = EnvVar{Secret: "secret", Value: "zza"}

	if e1 == e2 {
		t.Errorf("Expected %+v to be no equal to %+v, got true", e1, e2)
	}

	e1 = EnvVar{Secret: "secret", Value: "zzz"}
	e2 = EnvVar{Name: "a", Value: "2"}

	if e1 == e2 {
		t.Errorf("Expected %+v to be no equal to %+v, got true", e1, e2)
	}
}

func testFindByNameExist(t *testing.T) {
	var svc = Services{
		{Name: "ads"},
		{Name: "vpd"},
		{Name: "ooo"},
	}

	if svc.FindByName("vpd") == nil {
		t.Errorf("Expected service, got nil")
	}
}

func testFindByNameNotFound(t *testing.T) {
	var svc = Services{
		{Name: "ads"},
		{Name: "vpd"},
		{Name: "ooo"},
	}

	if s := svc.FindByName("aaa"); s != nil {
		t.Errorf("Expected nil, got %v", s)
	}
}

func TestServiceSortInterface(t *testing.T) {
	var s = Services{
		{Name: "b"},
		{Name: "a"},
		{Name: "c"},
	}

	var expected = Services{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}

	sort.Sort(s)
	if !reflect.DeepEqual(s, expected) {
		t.Errorf("Service sort invalid, got %q", s)
	}
}
