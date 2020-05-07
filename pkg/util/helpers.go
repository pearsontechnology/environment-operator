package util

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// LogTraceAsYaml logs anything as YAML when trace logging is enabled
func LogTraceAsYaml(identifier string, i interface{}) {
	cfgYAML, err := yaml.Marshal(&i)
	if err != nil {
		log.Fatalf("Error parsing YAML with id %s: %s", identifier, err.Error())
	}
	log.Tracef("%s: %#v", identifier, string(cfgYAML))
}
