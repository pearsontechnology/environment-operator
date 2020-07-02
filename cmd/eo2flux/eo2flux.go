package main

import (
	"fmt"
	"log"
	"os"
	"github.com/jessevdk/go-flags"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/flux"
)

type options struct {
	// Environment .bitesize file to read
	InputFile string `short:"i" long:"input-file" description:"input .bitesize file" required:"true"`

	// Where to write HelmRelease files
	OutputDir string `short:"o" long:"output-directory" description:"output directory for HelmRelease files" required:"true"`

	// Docker Image base path
	RegistryPath string `short:"r" long:"registry-path" description:"Docker registry path to inject in HelmRelease image: tags" default:"410908411653.dkr.ecr.us-east-1.amazonaws.com/glp2" required:"true"`

	// Consul  values file to read
	ConsulinFile string `short:"c" long:"Consule value file" description:"input consule value file"`
}

var opts options

var helmReleases map[string]string

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

	// Check whether there is any Consul export file is given as an input and load the contents
        if opts.ConsulinFile != "" {
		consulFile, err := flux.ConsulRead(opts.ConsulinFile)
	        helmReleases = flux.RenderHelmReleasesWithConsul(environmentFile, opts.RegistryPath, consulFile)
		if err != nil {
			panic(err)
		}
        }

	if err != nil {
		panic(err)
	}


	// retrieve dict of name: HelmRelease string
        if opts.ConsulinFile == "" {
		helmReleases = flux.RenderHelmReleases(environmentFile, opts.RegistryPath)
	}
	// write each HelmRelease to file
	for envSvcName, helmrelease := range helmReleases {
		outputFileName := fmt.Sprintf("%s/%s.yaml", opts.OutputDir, envSvcName)
		log.Printf("Writing %s\n", outputFileName)
		f, err := os.Create(outputFileName)
		if err != nil {
			panic(err)
		}
		f.WriteString(helmrelease)
		defer f.Close()
	}
}
