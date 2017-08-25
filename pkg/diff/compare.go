package diff

import (
	"github.com/kylelemons/godebug/pretty"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
)

// Compare returns unified diff between two bitesize Environment configs
// in a string
func Compare(config1, config2 bitesize.Environment) string {
	c1 := config1
	c2 := config2
	// Following fields are ignored for diff purposes
	c1.Tests = []bitesize.Test{}
	c2.Tests = []bitesize.Test{}
	c1.Deployment = nil
	c2.Deployment = nil
	c1.Name = ""
	c2.Name = ""

	// XXX: remove tprs for now
	var newServices bitesize.Services
	for _, s := range c1.Services {
		// if s.Type == "" {
		d := c2.Services.FindByName(s.Name)

		if d != nil {
			if s.Version == "" {
				s.Version = d.Version
			}
			s.Status = d.Status
			s.Deployment = d.Deployment
		}
		// Ignore diff for annotations if d.Deployment is nil
		if d == nil || d.Deployment == nil {
			s.Annotations = nil
		}

		newServices = append(newServices, s)
		// }
	}
	c1.Services = newServices

	// Ignore diff between service replicas when hpa is configured
	var correctedServices bitesize.Services
	gitServices := c1.Services
	clusterServices := c2.Services
	for _, gs := range gitServices {
		cs := clusterServices.FindByName(gs.Name)
		if cs != nil {
			if cs.HPA.MinReplicas != 0 {
				gs.Replicas = cs.Replicas
			}
		}
		correctedServices = append(correctedServices, gs)
	}
	c1.Services = correctedServices

	compareConfig := &pretty.Config{
		Diffable:       true,
		SkipZeroFields: true,
	}
	return compareConfig.Compare(c2, c1)
}
