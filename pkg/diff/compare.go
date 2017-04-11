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

	// XXX: remove tprs for now
	var newServices bitesize.Services
	for _, s := range c1.Services {
		// if s.Type == "" {
		d := c2.Services.FindByName(s.Name)

		if d != nil {
			s.Version = d.Version
			s.Status = d.Status
			s.Deployment = d.Deployment
		}
		newServices = append(newServices, s)
		// }
	}
	c1.Services = newServices
	// XXX: the end

	compareConfig := &pretty.Config{
		Diffable:       true,
		SkipZeroFields: true,
	}
	return compareConfig.Compare(c2, c1)
}
