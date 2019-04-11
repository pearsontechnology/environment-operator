package diff

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/godebug/pretty"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Compare creates a changeMap for the diff between environment configs and returns a boolean if changes were detected
func Compare(config1, config2 bitesize.Environment) bool {
	newChangeMap()

	c1 := config1 //New Config
	c2 := config2 //Existing Config

	// Following fields are ignored for diff purposes
	c1.Tests = []bitesize.Test{}
	c2.Tests = []bitesize.Test{}
	c1.Deployment = nil
	c2.Deployment = nil
	c1.Name = ""
	c2.Name = ""

	compareConfig := &pretty.Config{
		Diffable:          true,
		SkipZeroFields:    true,
		IncludeUnexported: false,
	}

	for _, s := range c1.Services {
		d := c2.Services.FindByName(s.Name)

		// Force changes for blue green "parent" service
		if d == nil && s.IsBlueGreenParentDeployment() {
			addServiceChange(s.Name, fmt.Sprintf("Name: +%s", s.Name))
		}

		// Ignore changes for active blue green deployment
		if s.IsActiveBlueGreenDeployment() {
			continue
		}

		if s.IsBlueGreenParentDeployment() {
			if d == nil {
				addServiceChange(s.Name, compareConfig.Compare(nil, s))
				continue
			}
			if serviceDiff := compareConfig.Compare(d.ActiveDeploymentName(), s.ActiveDeploymentName()); serviceDiff != "" {
				log.Debugf("Change detected for blue/green service %s - %s", s.Name, serviceDiff)
				addServiceChange(s.Name, serviceDiff)
				continue
			}
		}

		// compare configs only if deployment is found in cluster
		// and git service has no version set
		if (s.Version != "") || (d != nil && d.Version != "") {
			if d != nil {
				alignServices(&s, d)
			}

			if serviceDiff := compareConfig.Compare(d, s); serviceDiff != "" {
				log.Debugf("change detected for service %s - %s", s.Name, serviceDiff)
				addServiceChange(s.Name, serviceDiff)
			}
		}
	}
	return len(changeMap) > 0
}

// Can't think of a better word
func alignServices(src, dest *bitesize.Service) {
	//Note: src=new config    dest=existing config

	// Copy version from dest if source version is empty
	if src.Version == "" {
		src.Version = dest.Version
	}

	if src.Application == "" && dest.Application != "" {
		src.Application = dest.Application
	}

	// Copy status from dest (status is only stored in the cluster)
	src.Status = dest.Status

	// Ignore changes to internal info
	if src.Deployment != nil {
		src.Deployment.BlueGreen = nil
	}
	if dest.Deployment != nil {
		dest.Deployment.BlueGreen = nil
	}

	//If its a TPR type service, sync up the Limits since they aren't appied to the k8s resource
	if src.Type != "" {
		src.Limits.Memory = dest.Limits.Memory
		src.Limits.CPU = dest.Limits.CPU

	}

	//Sync up Requests in the case where different units are present, but they represent equivalent quantities
	destmemreq, _ := resource.ParseQuantity(dest.Requests.Memory)
	srcmemreq, _ := resource.ParseQuantity(src.Requests.Memory)
	destcpureq, _ := resource.ParseQuantity(dest.Requests.CPU)
	srccpureq, _ := resource.ParseQuantity(src.Requests.CPU)
	if destmemreq.Cmp(srcmemreq) == 0 {
		src.Requests.Memory = dest.Requests.Memory
	}
	if destcpureq.Cmp(srccpureq) == 0 {
		src.Requests.CPU = dest.Requests.CPU
	}

	//Sync up Limits in the case where different units are present, but they represent equivalent quantities
	destmemlim, _ := resource.ParseQuantity(dest.Limits.Memory)
	srcmemlim, _ := resource.ParseQuantity(src.Limits.Memory)
	destcpulim, _ := resource.ParseQuantity(dest.Limits.CPU)
	srccpulim, _ := resource.ParseQuantity(src.Limits.CPU)
	if destmemlim.Cmp(srcmemlim) == 0 {
		src.Limits.Memory = dest.Limits.Memory
	}
	if destcpulim.Cmp(srccpulim) == 0 {
		src.Limits.CPU = dest.Limits.CPU
	}

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
}
