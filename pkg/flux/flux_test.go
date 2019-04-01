package flux

import (
	"testing"
	"log"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/weaveworks/flux/integrations/apis/flux.weave.works/v1beta1"
	"gopkg.in/yaml.v2"
)

func TestRenderHelmReleases(t *testing.T) {
	e, err := bitesize.LoadFromFile("../../test/assets/environments.bitesize")
	if err != nil {
		t.Errorf("Unexpected error loading environment: %s", err.Error())
	}

	x := RenderHelmReleases(e, "docker.io")
	for k, v := range(x) {
		t := v1beta1.HelmRelease{}
		err := yaml.Unmarshal([]byte(v), &t)
		if err != nil {
				log.Fatalf("error in file %s: %v", k, err)
		}	
	}

}
