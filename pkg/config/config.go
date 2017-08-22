package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

// Config contains environment variables used to configure the app
type Config struct {
	LogLevel          string `envconfig:"LOG_LEVEL"`
	UseAuth           bool   `envconfig:"USE_AUTH" default:true`
	GitRepo           string `envconfig:"GIT_REMOTE_REPOSITORY"`
	GitBranch         string `envconfig:"GIT_BRANCH" default:"master"`
	GitKey            string `envconfig:"GIT_PRIVATE_KEY"`
	GitKeyPath        string `envconfig:"GIT_PRIVATE_KEY_PATH" default:"/etc/git/key"`
	GitLocalPath      string `envconfig:"GIT_LOCAL_PATH" default:"/tmp/repository"`
	EnvName           string `envconfig:"ENVIRONMENT_NAME"`
	EnvFile           string `envconfig:"BITESIZE_FILE"`
	Namespace         string `envconfig:"NAMESPACE"`
	DockerRegistry    string `envconfig:"DOCKER_REGISTRY" default:"bitesize-registry.default.svc.cluster.local:5000"`
	DockerPullSecrets string `envconfig:"DOCKER_PULL_SECRETS"`
	// AUTH stuff
	OIDCIssuerURL     string `envconfig:"OIDC_ISSUER_URL"`
	OIDCCAFile        string `envconfig:"OIDC_CA_FILE"`
	OIDCAllowedGroups string `envconfig:"OIDC_ALLOWED_GROUPS"`
	OIDCClientID      string `envconfig:"OIDC_CLIENT_ID" default:"bitesize"`

	HPAMaxReplicas int `envconfig:"HPA_MAX_REPLICAS" default:"50"`
	ReqMaxCPU      int `envconfig:"REQUESTS_MAX_CPU" default:"500"`

	TokenFile string `envconfig:"AUTH_TOKEN_FILE"`

	Debug string `envconfig:"DEBUG"`
}

var Env Config

func init() {
	err := envconfig.Process("operator", &Env)
	if err != nil {
		log.Fatal(err.Error())
	}

	if Env.LogLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}
}
