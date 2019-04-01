package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/flux"
	"log"
	"os"
)

type options struct {
	// Environment .bitesize file to read
	InputFile string `short:"i" long:"input-file" description:"input .bitesize file" required:"true"`

	// Where to write HelmRelease files
	OutputDir string `short:"o" long:"output-directory" description:"output directory for HelmRelease files" required:"true"`

	// Docker Image base path
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

	environmentFile, err := bitesize.LoadFromFile(opts.InputFile)
	if err != nil {
		panic(err)
	}
	for _, env := range environmentFile.Environments {
		for _, svc := range env.Services {

			if svc.Type == "" { // EO uses nil type for web apps
				svc.Type = "webservice"
			}

			outputFileName := fmt.Sprintf("%s/%s-%s.yaml",
				opts.OutputDir, env.Namespace, svc.Name)

			log.Printf("Writing %s\n", outputFileName)
			f, err := os.Create(outputFileName)
			if err != nil {
				panic(err)
			}
			s, err := flux.RenderHelmRelease(env, svc, opts.RegistryPath)
			if err != nil {
				panic(err)
			}
			f.WriteString(s)
			defer f.Close()
		}
	}
}
