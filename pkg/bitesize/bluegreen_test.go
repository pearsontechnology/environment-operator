package bitesize

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestInactiveBlueGreenDeployment(t *testing.T) {
	var saTests = []struct {
		Active   BlueGreenServiceSet
		Expected BlueGreenServiceSet
	}{
		{Active: BlueService, Expected: GreenService},
		{Active: GreenService, Expected: BlueService},
	}

	for _, test := range saTests {
		svc := &Service{
			Name:       "svc",
			Deployment: &DeploymentSettings{Mode: "bluegreen", BlueGreen: &BlueGreenSettings{Active: &test.Active}},
		}

		if svc.InactiveDeploymentTag() != test.Expected {
			t.Errorf("Unexpected inactive deployment, expected: %s, got: %s", test.Expected, svc.InactiveDeploymentName())
		}

	}
}

func TestBlueGreenURLForKind(t *testing.T) {
	var saTests = []struct {
		OriginalURL string
		ExpectedURL string
		Kind        BlueGreenServiceSet
	}{
		{OriginalURL: "www.some.url", ExpectedURL: "www-blue.some.url", Kind: BlueService},
		{OriginalURL: "www", ExpectedURL: "www-green", Kind: GreenService},
	}
	for _, test := range saTests {
		n := BlueGreenURLForKind(test.OriginalURL, test.Kind)
		if n != test.ExpectedURL {
			t.Errorf("Unexpected URL: expected %s, got %s", test.ExpectedURL, n)
		}
	}
}

func TestCopyBlueGreenService(t *testing.T) {
	svc := Service{}
	str := `
  name: test-service
  deployment:
    method: bluegreen
    active: blue
    custom_urls:
      blue:
      - www.blue.url
      green:
      - www.green.url
`

	if err := yaml.Unmarshal([]byte(str), &svc); err != nil {
		t.Fatalf("could not unmarshal yaml: %s", err.Error())
	}

	if svc.ActiveDeploymentName() != "test-service-blue" {
		t.Errorf("unexpected active deployment: expected test-service-blue, got %s", svc.ActiveDeploymentName())
	}

	blueService := copyBlueGreenService(svc, BlueService)
	if blueService.Name != "test-service-blue" {
		t.Errorf("Unexpected service name: expected test-service-blue, got %s", blueService.Name)
	}

	if !blueService.IsActiveBlueGreenDeployment() {
		t.Errorf("unexpected active blue green environment. expected true, got false")
	}

	if len(blueService.ExternalURL) != 1 {
		t.Errorf("unexpected urls for blue service url count: %d", len(blueService.ExternalURL))
	}

	if blueService.ExternalURL[0] != "www.blue.url" {
		t.Errorf("unexpected url for blue service: expected www.blue.url, got %s", blueService.ExternalURL[0])
	}

}
