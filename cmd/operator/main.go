package main

import (
	"math/rand"
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

var gitClient *git.Git
var client *cluster.Cluster
var reap reaper.Reaper

func init() {
	var err error

	rand.Seed(time.Now().UnixNano())

	gitClient = git.Client()

	client, err = cluster.Client()
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %s", err.Error())
	}

	reap = reaper.Reaper{
		Namespace: config.Env.Namespace,
		Wrapper:   client,
	}

	if config.Env.Debug != "" {
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

	err := gitClient.Pull()
	if err != nil {
		log.Errorf("Git clone error: %s", err.Error())
		log.Errorf("Git Client Information: \n RemotePath=%s \n LocalPath=%s \n Branch=%s \n SSHkey= \n %s", gitClient.RemotePath, gitClient.LocalPath, gitClient.BranchName, gitClient.SSHKey)
	}

	for {
		gitClient.Refresh()
		gitConfiguration, _ := bitesize.LoadEnvironmentFromConfig(config.Env)
		client.ApplyIfChanged(gitConfiguration)

		go reap.Cleanup(gitConfiguration)
		time.Sleep(30000 * time.Millisecond)
	}

}
