package diff

import (
	"testing"

	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
)

func TestDiffEmpty(t *testing.T) {
	a := bitesize.Environment{}
	b := bitesize.Environment{}

	if Compare(a, b) {
		t.Errorf("Expected diff to be empty, got: %s", Changes())
	}

}

func TestIgnoreTestFields(t *testing.T) {
	a := bitesize.Environment{Name: "E", Tests: []bitesize.Test{}}
	b := bitesize.Environment{Name: "A", Tests: []bitesize.Test{
		{Name: "a"},
	}}

	if Compare(a, b) {
		t.Errorf("Expected diff to be empty, got: %s", Changes())
	}
}

func TestIgnoreDeploymentFields(t *testing.T) {
	a := bitesize.Environment{Deployment: &bitesize.DeploymentSettings{}}
	b := bitesize.Environment{Deployment: &bitesize.DeploymentSettings{
		Method: "bluegreen",
	}}

	if Compare(a, b) {
		t.Errorf("Expected diff to be empty, got: %s", Changes())
	}
}

func TestIgnoreStatusFields(t *testing.T) {
	a := bitesize.Environment{
		Services: bitesize.Services{
			{
				Name: "a",
				Status: bitesize.ServiceStatus{
					AvailableReplicas: 3,
				},
			},
		},
	}

	b := bitesize.Environment{
		Services: bitesize.Services{
			{
				Name: "a",
				Status: bitesize.ServiceStatus{
					AvailableReplicas: 1,
				},
			},
		},
	}

	if Compare(a, b) {
		t.Errorf("Expected diff to be empty, got: %s", Changes())
	}
}

func TestDiffNames(t *testing.T) {
	a := bitesize.Environment{Name: "asd"}
	b := bitesize.Environment{Name: "asdf"}

	if Compare(a, b) {
		t.Error("Expected diff, got the same")
	}
}

func TestQuantitiesWithDiffUnits(t *testing.T) {
	a := bitesize.Environment{
		Services: bitesize.Services{
			{
				Name: "a",
				Requests: bitesize.ContainerRequests{
					CPU:    "1000m",
					Memory: "2048Mi",
				},
				Limits: bitesize.ContainerLimits{
					CPU:    "1000m",
					Memory: "2048Mi",
				},
			},
		},
	}

	b := bitesize.Environment{
		Services: bitesize.Services{
			{
				Name: "a",
				Requests: bitesize.ContainerRequests{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: bitesize.ContainerLimits{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}

	if Compare(a, b) {
		t.Errorf("Expected to be the same, but got diff %s", Changes())
	}
}

func TestBlueGreenExternalUrls(t *testing.T) {
	var saTests = []struct {
		versionA []string
		versionB []string
		expected bool
	}{
		{
			[]string{"www.a.com"},
			[]string{"www.b.com"},
			true,
		},
		{
			[]string{"www.a.com"},
			[]string{"www.a.com"},
			false,
		},
		{
			[]string{"www.a.com"},
			[]string{},
			true,
		},
		{
			[]string{},
			[]string{"www.a.com"},
			true,
		},
		{
			[]string{"www.a.com", "www.b.com"},
			[]string{"www.a.com"},
			true,
		},
		{
			[]string{"www.a.com", "www.b.com"},
			[]string{"www.a.com", "www.b.com"},
			false,
		},
	}

	for _, tst := range saTests {
		a := bitesize.Environment{
			Name: "e",
			Services: []bitesize.Service{
				{
					Name:        "a",
					ExternalURL: tst.versionA,
					Deployment: &bitesize.DeploymentSettings{
						Method: "bluegreen",
					},
				},
			},
		}
		b := bitesize.Environment{
			Name: "e",
			Services: []bitesize.Service{
				{
					Name:        "a",
					ExternalURL: tst.versionB,
					Deployment: &bitesize.DeploymentSettings{
						Method: "bluegreen",
					},
				},
			},
		}

		if Compare(a, b) != tst.expected {
			t.Errorf(
				"Unexpected external url compare(%q, %q), should be %t, got %t \n%+v",
				tst.versionA, tst.versionB, tst.expected, Compare(a, b), Changes(),
			)
		}
	}
}

func TestDiffVersionsSame(t *testing.T) {
	var saTests = []struct {
		versionA string
		versionB string
		expected bool
	}{
		{"1", "1", false},
		{"1", "2", true},
		{"", "1", false}, // assume the same if environments.bitesize does not have version
		{"1", "", true},  // assume diff  if cluster is not deployed
	}

	for _, tst := range saTests {
		a := bitesize.Environment{
			Name: "a", Services: []bitesize.Service{{Name: "a", Version: tst.versionA}},
		}
		b := bitesize.Environment{
			Name: "a", Services: []bitesize.Service{{Name: "a", Version: tst.versionB}},
		}

		if Compare(a, b) != tst.expected {
			t.Errorf(
				"Unexpected version compare(%s,%s) should be %t\n%s\n A %+v\n B %+v",
				tst.versionA, tst.versionB, tst.expected, Changes(), a.Services, b.Services,
			)
		}
	}
}
