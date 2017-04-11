package bitesize

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

// Tests to see that YAML documents unmarshal correctly

func TestUnmarshalPorts(t *testing.T) {
	t.Run("ports string parsed correctly", testPortsString)
	t.Run("ports preferred over port", testPortsOverPort)
	t.Run("ports with invalid value", testPortsWithInvalidValue)
	t.Run("empty ports return default", testPortsEmpty)
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

	if !eqIntArrays(svc.Ports, []int{81, 88, 89}) {
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

	if !eqIntArrays(svc.Ports, []int{81, 82}) {
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

	if !eqIntArrays(svc.Ports, []int{80}) {
		t.Errorf("Unexpected ports: %v", svc.Ports)
	}
}

func eqIntArrays(a, b []int) bool {

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
