package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

// Config contains environment variables used to configure the app
type Config struct {
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	UseAuth      bool   `envconfig:"USE_AUTH" default:"true"`
	GitRepo      string `envconfig:"GIT_REMOTE_REPOSITORY"`
	GitBranch    string `envconfig:"GIT_BRANCH" default:"master"`
	GitKey       string `envconfig:"GIT_PRIVATE_KEY"`
	GitKeyPath   string `envconfig:"GIT_PRIVATE_KEY_PATH" default:"/etc/git/key"`
	GitUser      string `envconfig:"GIT_USER"`
	GitToken     string `envconfig:"GIT_TOKEN"`
	GitLocalPath string `envconfig:"GIT_LOCAL_PATH" default:"/tmp/repository"`
	GitRootPath  string `envconfig:"GIT_ROOT_PATH" default:"/tmp/"`

	//Gists
	GistsUser  string `envconfig:"GISTS_USER"`
	GistsToken string `envconfig:"GISTS_TOKEN"`
	GistsKey   string `envconfig:"GISTS_PRIVATE_KEY"`

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

	HPAMaxReplicas     int    `envconfig:"HPA_MAX_REPLICAS" default:"50"`
	LimitMaxCPU        int    `envconfig:"LIMITS_MAX_CPU" default:"4000"`          //4 Cores
	LimitMaxMemory     int    `envconfig:"LIMITS_MAX_MEMORY" default:"8192"`       //8Gib
	LimitDefaultCPU    string `envconfig:"LIMITS_DEFAULT_CPU" default:"1000m"`     //1 Core
	LimitDefaultMemory string `envconfig:"LIMITS_DEFAULT_MEMORY" default:"2048Mi"` //2Gib
	RequestsDefaultCPU string `envconfig:"REQUESTS_DEFAULT_CPU" default:"100m"`

	TokenFile string `envconfig:"AUTH_TOKEN_FILE"`

	Debug string `envconfig:"DEBUG"`
}

// Env parses and exports configuration for
// operator
var Env Config

func init() {
	err := envconfig.Process("operator", &Env)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ensure only a single type of auth is used.
	if Env.GitKey != "" && Env.GitToken != "" {
		log.Fatal("Please choose either Gitkey or GitToken but not both")
	}

	logLevel, err := log.ParseLevel(Env.LogLevel)
	if err != nil {
		log.Fatalf("Can't set loglevel \"%s\": %s", Env.LogLevel, err.Error())
	}
	log.SetLevel(logLevel)
}
