package flux

import (
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/gobuffalo/packr/v2"
	"fmt"
	"bytes"
	"text/template"
)

var box *packr.Box

func init() {
	box = packr.New("eo2flux", "./templates")
}

// RenderHelmRelease creates a Weaveworks Flux-compatible "HelmRelease" file
func RenderHelmRelease(env bitesize.Environment, svc bitesize.Service, registryPath string) (string, error) {
	templateFile := ""
	switch svc.Type {
	case "webservice":
		templateFile = "webservice.tmpl"
	case "Cb":
		templateFile = "couchbase.tmpl"
	case "Kafka":
		templateFile = "kafka.tmpl"
	case "Zookeeper":
		templateFile = "zookeeper.tmpl"
	case "postgres":
		templateFile = "postgres.tmpl"
	case "mysql":
		templateFile = "mysql.tmpl"
	case "Neptune":
		templateFile = "neptune.tmpl"
	default:
		return "", fmt.Errorf("Couldn't determine template file for type \"%s\"", svc.Type)
	}

	templateText, err := box.FindString(templateFile)
	if err != nil {
		return "", fmt.Errorf("No template available for type \"%s\"", err.Error())
	}

	t := template.New("T")
	tmpl, err := t.Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("Error opening file \"%s\": %s", templateFile, err.Error())
	}

	data := struct {
		Environment bitesize.Environment
		Service     bitesize.Service
		Registry    string
	}{
		env,
		svc,
		registryPath,
	}
	var rendered bytes.Buffer
	tmpl.Execute(&rendered, data)

	return rendered.String(), nil
}
