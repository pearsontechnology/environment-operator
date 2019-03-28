package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"log"
	"os"
	"text/template"
)

func main() {
	inputFile := flag.String("i", "./environment.sample", "input .bitesize file")
	outputDir := flag.String("o", "./helmreleases", "output directory for HelmRelease files")
	registryPath := flag.String("r", "815492460363.dkr.ecr.us-east-1.amazonaws.com/glp2", "docker registry path for app image:")
	flag.Parse()
	// load templates
	box := packr.New("eo2flux", "./templates")

	environmentFile, err := bitesize.LoadFromFile(*inputFile)
	if err != nil {
		panic(err)
	}
	for _, environment := range environmentFile.Environments {
		for _, service := range environment.Services {

			if service.Type == "" { // EO uses nil type for web apps
				service.Type = "webservice"
			}
			
			outputFileName := fmt.Sprintf("%s/%s-%s.yaml",
				*outputDir, environment.Namespace, service.Name)

			log.Printf("Writing %s\n", outputFileName)
			f, err := os.Create(outputFileName)
			if err != nil {
				panic(err)
			}
			f.WriteString(renderHelmRelease(box, environment, service, registryPath))
			defer f.Close()

		}
	}
}

func renderHelmRelease(box *packr.Box, env bitesize.Environment, svc bitesize.Service, registryPath *string) string {
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
		log.Fatalf("Couldn't determine template file for type \"%s\"", svc.Type)
	}

	templateText, err := box.FindString(templateFile)
	if err != nil {
		log.Fatalf("No template available for type \"%s\"", err.Error())
	}

	t := template.New("T")
	tmpl, err := t.Parse(templateText)
	if err != nil {
		log.Fatalf("Error opening file \"%s\": %s", templateFile, err.Error())
	}

	data := struct {
		Environment bitesize.Environment
		Service     bitesize.Service
		Registry    string
	}{
		env,
		svc,
		*registryPath,
	}
	var rendered bytes.Buffer
	tmpl.Execute(&rendered, data)

	return rendered.String()
}
