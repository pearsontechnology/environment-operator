package diff

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/godebug/pretty"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Compare creates a changeMap for the diff between environment configs and returns a boolean if changes were detected
func Compare(desiredCfg, existingCfg bitesize.Environment) bool {
	newChangeMap()

	// Following fields are ignored for diff purposes
	desiredCfg.Tests = []bitesize.Test{}
	existingCfg.Tests = []bitesize.Test{}
	desiredCfg.Deployment = nil
	existingCfg.Deployment = nil
	desiredCfg.Name = ""
	existingCfg.Name = ""

	compareConfig := &pretty.Config{
		Diffable:          true,
		SkipZeroFields:    true,
		IncludeUnexported: false,
	}

	for _, desiredCfgSvc := range desiredCfg.Services {
		existingCfgSvc := existingCfg.Services.FindByName(desiredCfgSvc.Name)

		// Force changes for blue green "parent" service
		if existingCfgSvc == nil && desiredCfgSvc.IsBlueGreenParentDeployment() {
			addServiceChange(desiredCfgSvc.Name, fmt.Sprintf("Name: +%s", desiredCfgSvc.Name))
		}

		// Ignore changes for active blue green deployment
		if desiredCfgSvc.IsActiveBlueGreenDeployment() {
			continue
		}

		if desiredCfgSvc.IsBlueGreenParentDeployment() {
			if existingCfgSvc == nil {
				addServiceChange(desiredCfgSvc.Name, compareConfig.Compare(nil, desiredCfgSvc))
				continue
			}
			if serviceDiff := compareConfig.Compare(existingCfgSvc.ActiveDeploymentName(), desiredCfgSvc.ActiveDeploymentName()); serviceDiff != "" {
				log.Debugf("Change detected for blue/green service %s - %s", desiredCfgSvc.Name, serviceDiff)
				addServiceChange(desiredCfgSvc.Name, serviceDiff)
				continue
			}
		}

		// compare configs only if deployment is found in cluster
		// and git service has no version set
		if (desiredCfgSvc.Version != "") || (existingCfgSvc != nil && existingCfgSvc.Version != "") {
			if existingCfgSvc != nil {
				alignServices(&desiredCfgSvc, existingCfgSvc)
			}

			if serviceDiff := compareConfig.Compare(existingCfgSvc, desiredCfgSvc); serviceDiff != "" {
				log.Debugf("Change detected for service %s - %s", desiredCfgSvc.Name, serviceDiff)
				addServiceChange(desiredCfgSvc.Name, serviceDiff)
			}
		}
	}
	return len(changeMap) > 0
}

// Can't think of a better word
func alignServices(desiredCfg, currentCfg *bitesize.Service) {

	// Copy version from currentCfg if source version is empty
	if desiredCfg.Version == "" {
		desiredCfg.Version = currentCfg.Version
	}

	if desiredCfg.Application == "" && currentCfg.Application != "" {
		desiredCfg.Application = currentCfg.Application
	}

	// Copy status from currentCfg (status is only stored in the cluster)
	desiredCfg.Status = currentCfg.Status

	// Ignore changes to internal info
	if desiredCfg.Deployment != nil {
		desiredCfg.Deployment.BlueGreen = nil
	}
	if currentCfg.Deployment != nil {
		currentCfg.Deployment.BlueGreen = nil
	}

	//If its a TPR type service, sync up the Limits since they aren't appied to the k8s resource
	if desiredCfg.Type != "" {
		desiredCfg.Limits.Memory = currentCfg.Limits.Memory
		desiredCfg.Limits.CPU = currentCfg.Limits.CPU

	}

	//Sync up Requests in the case where different units are present, but they represent equivalent quantities
	destmemreq, _ := resource.ParseQuantity(currentCfg.Requests.Memory)
	srcmemreq, _ := resource.ParseQuantity(desiredCfg.Requests.Memory)
	destcpureq, _ := resource.ParseQuantity(currentCfg.Requests.CPU)
	srccpureq, _ := resource.ParseQuantity(desiredCfg.Requests.CPU)
	if destmemreq.Cmp(srcmemreq) == 0 {
		desiredCfg.Requests.Memory = currentCfg.Requests.Memory
	}
	if destcpureq.Cmp(srccpureq) == 0 {
		desiredCfg.Requests.CPU = currentCfg.Requests.CPU
	}

	//Sync up Limits in the case where different units are present, but they represent equivalent quantities
	destmemlim, _ := resource.ParseQuantity(currentCfg.Limits.Memory)
	srcmemlim, _ := resource.ParseQuantity(desiredCfg.Limits.Memory)
	destcpulim, _ := resource.ParseQuantity(currentCfg.Limits.CPU)
	srccpulim, _ := resource.ParseQuantity(desiredCfg.Limits.CPU)
	if destmemlim.Cmp(srcmemlim) == 0 {
		desiredCfg.Limits.Memory = currentCfg.Limits.Memory
	}
	if destcpulim.Cmp(srccpulim) == 0 {
		desiredCfg.Limits.CPU = currentCfg.Limits.CPU
	}

	// Override source replicas with currentCfg replicas if HPA is active
	if currentCfg.HPA.MinReplicas != 0 {
		desiredCfg.Replicas = currentCfg.Replicas
	}

	if currentCfg.Version == "" {
		// If no deployment yet, ignore annotations. They only apply onto
		// deployment object.
		desiredCfg.Annotations = currentCfg.Annotations
	} else {
		// Apply all existing annotations
		for k, v := range currentCfg.Annotations {
			if desiredCfg.Annotations[k] == "" {
				desiredCfg.Annotations[k] = v
			}
		}
	}
}
