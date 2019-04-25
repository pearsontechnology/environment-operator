package bitesize

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"
)

// BlueGreenServiceSet is a type specifying whether environment is blue or green
type BlueGreenServiceSet int

const (
	BlueService  BlueGreenServiceSet = 1
	GreenService BlueGreenServiceSet = 2
)

var bgToString = map[BlueGreenServiceSet]string{
	BlueService:  "blue",
	GreenService: "green",
}

var bgToID = map[string]BlueGreenServiceSet{
	"blue":  BlueService,
	"green": GreenService,
}

// BlueGreenDeploymentID converts string representation to BlueGreenServiceSet type
func BlueGreenDeploymentID(name string) BlueGreenServiceSet {
	return bgToID[name]
}

func (e BlueGreenServiceSet) String() string {
	return bgToString[e]
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for BlueGreenServiceSet.
func (e *BlueGreenServiceSet) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}

	if s == "" {
		*e = BlueService
	}
	if bgToID[s] == 0 {
		return fmt.Errorf("blue_green: invalid type %s", s)
	}

	*e = bgToID[s]
	return nil
}

// BlueGreenURLForKind constructs preformatted URL for environment given
// "parent" service's url and environment's colour
func BlueGreenURLForKind(url string, kind BlueGreenServiceSet) string {
	split := strings.Split(url, ".")
	split[0] = fmt.Sprintf("%s-%s", split[0], kind)
	return strings.Join(split, ".")
}

// copyBlueGreenService creates a copy of current service with deployment
// set to rollingupgrade and a name suffixed with either -blue or -green
func copyBlueGreenService(svc Service, kind BlueGreenServiceSet) Service {
	retval := Service{}
	byt, err := json.Marshal(svc)
	if err != nil {
		log.Errorf("copy blue green service marshal error: %s", err.Error())
	}
	err = json.Unmarshal(byt, &retval)
	if err != nil {
		log.Errorf("copy blue green service unmarshal error: %s", err.Error())
	}

	retval.Name = fmt.Sprintf("%s-%s", svc.Name, kind)

	retval.Deployment = &DeploymentSettings{
		Method:    "rolling-upgrade",
		BlueGreen: &BlueGreenSettings{DeploymentColour: &kind},
	}

	if svc.ActiveDeploymentTag() == kind {
		retval.Deployment.BlueGreen.ActiveFlag = true
	}

	// If custom_urls set, use them as external_url. Otherwise, generate urls
	externalURLs := svc.Deployment.CustomURLs[kind.String()]
	if len(externalURLs) == 0 {
		for _, u := range svc.ExternalURL {
			externalURLs = append(externalURLs, BlueGreenURLForKind(u, kind))
		}
	}
	retval.ExternalURL = externalURLs

	return retval
}
