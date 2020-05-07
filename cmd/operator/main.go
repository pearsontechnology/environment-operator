package main

import (
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
	"github.com/pearsontechnology/environment-operator/pkg/reaper"
	"github.com/pearsontechnology/environment-operator/pkg/web"
	"github.com/pearsontechnology/environment-operator/version"
)

var gitClient *git.Git
var client *cluster.Cluster
var reap reaper.Reaper

func init() {
	var err error

	gitClient = git.Client()
	log.Tracef("gitClient: %#v", gitClient)

	client, err = cluster.Client()
	if err != nil {
		log.Fatalf("Error initializing kubernetes client: %s", err.Error())
	}

	reap = reaper.Reaper{
		Namespace: config.Env.Namespace,
		Wrapper:   client,
	}

	logLevel, err := log.ParseLevel(config.Env.LogLevel)
	if err != nil {
		log.Fatalf("Can't parse LOG_LEVEL \"%s\": %s", logLevel, err.Error())
	}
	log.SetLevel(logLevel)

	// allow DEBUG=true to override LOG_LEVEL
	if config.Env.Debug == "true" {
		log.SetLevel(log.DebugLevel)
	}
}

func webserver() {
	logged := handlers.CombinedLoggingHandler(os.Stderr, web.Router())
	authenticated := logged

	if config.Env.UseAuth {
		authenticated = web.Auth(logged)
	}

	if err := http.ListenAndServe(":8080", authenticated); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Infof("Starting up environment-operator version %s", version.Version)

	go webserver()

	// Polling interval
	sleepDurationSeconds := 30
	sleepDuration := time.Duration(sleepDurationSeconds) * time.Second

	err := gitClient.Pull()

	if err != nil {
		log.Errorf("Git clone error: %s", err.Error())
		log.Errorf("Git Client Information:")
		log.Errorf("  RemotePath=%s"+
			"   LocalPath=%s"+
			"   Branch=%s"+
			"   SSHkey=%s"+
			"   GITUser=%s\n",
			gitClient.RemotePath,
			gitClient.LocalPath,
			gitClient.BranchName,
			gitClient.SSHKey,
			gitClient.GitUser,
		)
	}

	for {
		if err := gitClient.Refresh(); err != nil {
			log.Errorf("git client refresh failed with %s", err.Error())
		}
		configurationInGit, err := bitesize.LoadEnvironmentFromConfig(config.Env)
		log.Tracef("configurationInGit: %#v", configurationInGit)

		if err != nil {
			log.Errorf("error while loading environment config: %s", err.Error())
		} else {
			if err := client.ApplyIfChanged(configurationInGit); err != nil {
				log.Errorf("error when applying changes: %s", err.Error())
			}
			if err := reap.Cleanup(configurationInGit); err != nil {
				log.Errorf("error reaper failed: %s", err.Error())
			}
		}
		log.Debugf("Sleeping %d seconds", sleepDurationSeconds)
		time.Sleep(sleepDuration)
	}

}
