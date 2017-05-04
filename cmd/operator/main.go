package main

import (
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/git"
	"github.com/pearsontechnology/environment-operator/pkg/reaper"
	"github.com/pearsontechnology/environment-operator/pkg/web"
	"github.com/pearsontechnology/environment-operator/version"

	"github.com/gorilla/handlers"
)

var cfg config.Config
var gitClient *git.Git
var client *cluster.Cluster
var reap reaper.Reaper

func init() {
	var err error

	cfg = config.Load()
	gitClient = git.Client()

	client, err = cluster.Client()
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %s", err.Error())
	}

	reap = reaper.Reaper{
		Namespace: cfg.Namespace,
		Wrapper:   client,
	}

	if cfg.Debug != "" {
		log.SetLevel(log.DebugLevel)
	}
}

func webserver() {
	logged := handlers.CombinedLoggingHandler(os.Stderr, web.Router())
	authenticated := logged

	if cfg.UseAuth {
		authenticated = web.Auth(logged)
	}

	if err := http.ListenAndServe(":8080", authenticated); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Infof("Starting up environment-operator version %s", version.Version)

	go webserver()

	gitClient.Clone()

	for {
		gitClient.Refresh()
		gitConfiguration, _ := bitesize.LoadEnvironmentFromConfig(cfg)
		client.ApplyIfChanged(gitConfiguration)

		go reap.Cleanup(gitConfiguration)
		time.Sleep(30000 * time.Millisecond)
	}

}
