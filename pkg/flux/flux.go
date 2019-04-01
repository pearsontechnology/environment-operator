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

// RenderHelmReleases creates a map of serviceIdentifier:HelmRelease yaml
func RenderHelmReleases(envs *bitesize.EnvironmentsBitesize, regPath string) map[string]string {
	m := make(map[string]string)

	for _, env := range envs.Environments {
		for _, svc := range env.Services {

			if svc.Type == "" { // EO uses nil type for web apps
				svc.Type = "webservice"
			}
			key := fmt.Sprintf("%s-%s", env.Namespace, svc.Name)

			val, err := RenderHelmRelease(env, svc, regPath)
			if err != nil {
				panic(err)
			}
			m[key] = val
		}
	}
	return m
}

// RenderHelmRelease creates a Weaveworks Flux-compatible "HelmRelease" file
func RenderHelmRelease(env bitesize.Environment, svc bitesize.Service, registryPath string) (string, error) {
	templateFile := fmt.Sprintf("%s.tmpl", svc.Type)

	templateText, err := box.FindString(templateFile)
	if err != nil {
		return "", fmt.Errorf("No template available for type \"%s\": %s", svc.Type, err.Error())
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
