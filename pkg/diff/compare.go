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

	var newServices bitesize.Services
	for _, s := range c1.Services {
		d := c2.Services.FindByName(s.Name)

		alignServices(&s, d)

		newServices = append(newServices, s)
	}
	c1.Services = newServices

	compareConfig := &pretty.Config{
		Diffable:       true,
		SkipZeroFields: true,
	}
	return compareConfig.Compare(c2, c1)
}

// Can't think of a better word
func alignServices(src, dest *bitesize.Service) {
	if dest != nil {
		// Copy version from dest if source version is empty
		if src.Version == "" {
			src.Version = dest.Version
		}

		// Copy status from dest (status is only stored in the cluster)
		src.Status = dest.Status

		// Override source replicas with dest replicas if HPA is active
		if dest.HPA.MinReplicas != 0 {
			src.Replicas = dest.Replicas
		}

		if dest.Version == "" {
			// If no deployment yet, ignore annotations. They only apply onto
			// deployment object.
			src.Annotations = dest.Annotations
		} else {
			// Apply all existing annotations
			for k, v := range dest.Annotations {
				if src.Annotations[k] == "" {
					src.Annotations[k] = v
				}
			}
		}

		// I don't like it. All of these should live somewhere else rather than in
		// diff logic :(
		if len(dest.DeployedPods) > 0 {
			src.DeployedPods = dest.DeployedPods
		}

	}

}
