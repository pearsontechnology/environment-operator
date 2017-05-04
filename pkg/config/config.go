package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

// Config contains environment variables used to configure the app
type Config struct {
	LogLevel       string `envconfig:"LOG_LEVEL"`
	UseAuth        bool   `envconfig:"USE_AUTH" default:true`
	GitRepo        string `envconfig:"GIT_REMOTE_REPOSITORY"`
	GitBranch      string `envconfig:"GIT_BRANCH" default:"master"`
	GitKey         string `envconfig:"GIT_PRIVATE_KEY"`
	GitLocalPath   string `envconfig:"GIT_LOCAL_PATH" default:"/tmp/repository"`
	EnvName        string `envconfig:"ENVIRONMENT_NAME"`
	EnvFile        string `envconfig:"BITESIZE_FILE"`
	Namespace      string `envconfig:"NAMESPACE"`
	DockerRegistry string `envconfig:"DOCKER_REGISTRY" default:"bitesize-registry.default.svc.cluster.local:5000"`
	// AUTH stuff
	OIDCIssuerURL     string `envconfig:"OIDC_ISSUER_URL"`
	OIDCCAFile        string `envconfig:"OIDC_CA_FILE"`
	OIDCAllowedGroups string `envconfig:"OIDC_ALLOWED_GROUPS"`
	OIDCClientID      string `envconfig:"OIDC_CLIENT_ID" default:"bitesize"`

	Debug string `envconfig:"DEBUG"`
}

func Load() Config {
	var cfg Config
	err := envconfig.Process("operator", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	if cfg.LogLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}
	return cfg
}
