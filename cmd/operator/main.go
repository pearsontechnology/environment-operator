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
		log.Errorf("Git Client Information: \n RemotePath=%s \n LocalPath=%s \n Branch=%s \n SSHkey=%s \n  GITUser= %s \n", gitClient.RemotePath, gitClient.LocalPath, gitClient.BranchName, gitClient.SSHKey, gitClient.GitUser)
	}

	for {
		gitClient.Refresh()
		gitConfiguration, err := bitesize.LoadEnvironmentFromConfig(config.Env)

		if err != nil {
			log.Errorf("error while loading environment config: %s", err.Error())
		} else {
			client.ApplyIfChanged(gitConfiguration)
			reap.Cleanup(gitConfiguration)
		}

		time.Sleep(30000 * time.Millisecond)
	}

}
