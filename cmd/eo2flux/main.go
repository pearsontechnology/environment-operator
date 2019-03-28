package main

import (
	"bytes"
	"github.com/jessevdk/go-flags"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"log"
	"os"
	"text/template"
)

type options struct {
	// Environment .bitesize file to read
	InputFile string `short:"i" long:"input-file" description:"input .bitesize file" required:"true"`

	// Where to write HelmRelease files
	OutputDir string `short:"o" long:"output-directory" description:"output directory for HelmRelease files" required:"true"`

	// Where to write HelmRelease files
	RegistryPath string `short:"r" long:"registry-path" description:"Docker registry path to inject in HelmRelease image: tags" default:"815492460363.dkr.ecr.us-east-1.amazonaws.com/glp2" required:"true"`
}

var opts options

var parser = flags.NewParser(&opts, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	// load templates
	box := packr.New("eo2flux", "./templates")

	environmentFile, err := bitesize.LoadFromFile(opts.InputFile)
	if err != nil {
		panic(err)
	}
	for _, environment := range environmentFile.Environments {
		for _, service := range environment.Services {

			if service.Type == "" { // EO uses nil type for web apps
				service.Type = "webservice"
			}
			
			outputFileName := fmt.Sprintf("%s/%s-%s.yaml",
			opts.OutputDir, environment.Namespace, service.Name)

			log.Printf("Writing %s\n", outputFileName)
			f, err := os.Create(outputFileName)
			if err != nil {
				panic(err)
			}
			f.WriteString(renderHelmRelease(box, environment, service, opts.RegistryPath))
			defer f.Close()

		}
	}
}

func renderHelmRelease(box *packr.Box, env bitesize.Environment, svc bitesize.Service, registryPath string) string {
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
		registryPath,
	}
	var rendered bytes.Buffer
	tmpl.Execute(&rendered, data)

	return rendered.String()
}
