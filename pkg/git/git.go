package git

import (
	"fmt"
	"net"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"golang.org/x/crypto/ssh"
	gogit "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// Git represents repository object and wraps go-git calls
type Git struct {
	SSHKey     string
	LocalPath  string
	RemotePath string
	BranchName string
	Repository *gogit.Repository
	GitToken   string
	GitUser    string
}

// Client initializes a git repo under a temp directory
// and attaches a remote
func Client() *Git {
	var repository *gogit.Repository
	var err error

	if _, err = os.Stat(config.Env.GitLocalPath); os.IsNotExist(err) {
		repository, err = gogit.PlainInit(config.Env.GitLocalPath, false)
		if err != nil {
			log.Errorf("could not init local repository %s: %s", config.Env.GitLocalPath, err.Error())
		}
	} else {
		repository, err = gogit.PlainOpen(config.Env.GitLocalPath)
	}

	if _, err = repository.Remote("origin"); err == gogit.ErrRemoteNotFound {
		_, err = repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{config.Env.GitRepo},
		})
		if err != nil {
			log.Errorf("could not attach to origin %s: %s", config.Env.GitRepo, err.Error())
		}
	}

	git := Git{
		LocalPath:  config.Env.GitLocalPath,
		RemotePath: config.Env.GitRepo,
		BranchName: config.Env.GitBranch,
		Repository: repository,
		SSHKey:     config.Env.GitKey,
	}

	if len(config.Env.GitToken) > 0 {
		log.Debug("using Git token")
		git.GitToken = config.Env.GitToken
		git.GitUser = config.Env.GitUser
		git.SSHKey = "" // just to make sure we set it empty, config.Env.GitKey default is ""
	}
	return &git
}

// EnvGitClient is git client for environment
func EnvGitClient(repo string, branch string, namespace string, env string) (*Git, error) {

	localPath := path.Join(config.Env.GitRootPath, namespace, env)

	var repository *gogit.Repository

	repository, err := gogit.PlainOpen(localPath)

	if err == gogit.ErrRepositoryNotExists {
		log.Debugf("environment %s repository does not exist initializing a new empty repository", env)
		repository, err = gogit.PlainInit(localPath, false)
		if err != nil {
			return nil, fmt.Errorf("could not init local repository %s: %s", localPath, err.Error())
		}
	}

	remote, err := repository.Remote("origin")
	if err == gogit.ErrRemoteNotFound {
		log.Debugf("remote not found, generating new origin")
		remote, err = repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{repo},
		})
		if err != nil {
			log.Errorf("could not attach to origin %s: %s", repo, err.Error())
		}
	}

	// remote has been changed re-init repo
	if remote == nil || remote.Config() == nil || remote.Config().URLs[0] != repo {
		log.Debugf("remote has been changed, re-initializing repository %s", repo)
		err := os.RemoveAll(localPath)
		if err != nil {
			log.Errorf("EnvGitClient re-init remove failed: %s", err.Error())
		}
		repository, err = gogit.PlainInit(localPath, false)
		if err != nil {
			return nil, fmt.Errorf("could not init local repository %s: %s", localPath, err.Error())
		}
		remote, err = repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{repo},
		})
		if err != nil {
			log.Errorf("could not attach to origin %s: %s", repo, err.Error())
		}
	}

	// if gists ssh key
	key := config.Env.GitKey
	token := config.Env.GitToken
	user := config.Env.GitUser

	// fallback to git configuration
	if len(config.Env.GistsKey) > 0 {
		key = config.Env.GistsKey
	}

	if len(config.Env.GistsToken) > 0 {
		token = config.Env.GistsToken
	}

	if len(config.Env.GistsUser) > 0 {
		token = config.Env.GistsUser
	}

	git := Git{
		LocalPath:  localPath,
		RemotePath: remote.Config().URLs[0],
		BranchName: branch,
		SSHKey:     key,
		Repository: repository,
	}

	if len(token) > 0 {
		log.Debug("using Git token")
		git.GitToken = token
		git.GitUser = user
		git.SSHKey = "" // just to make sure we set it empty, config.Env.GitKey default is ""
	}
	return &git, nil
}

// Setup options for git pull
func (g *Git) pullOptions() *gogit.PullOptions {
	branch := fmt.Sprintf("refs/heads/%s", g.BranchName)
	// Return options with token auth if enabled
	log.Debug("performing git pull")
	opt := gogit.PullOptions{
		ReferenceName: plumbing.ReferenceName(branch),
	}

	if config.Env.UseAuth {
		opt.Auth = g.auth()
	}

	return &opt
}

// Setup options for fetch. This will also be used
// for recording changes to git repo
func (g *Git) fetchOptions() *gogit.FetchOptions {
	//Return options with token auth if enabled
	log.Debug("performing git fetch")
	opt := gogit.FetchOptions{}

	if config.Env.UseAuth {
		opt.Auth = g.auth()
	}

	return &opt
}

// Auth returns AuthMethod object based on
// authentication mechanism chosen
func (g *Git) auth() transport.AuthMethod {

	if g.GitToken != "" {
		log.Debug("using gittoken")
		log.Debug(g.GitUser)
		return &githttp.BasicAuth{
			Username: g.GitUser,
			Password: g.GitToken,
		}
	}
	log.Debug("using sshauth")
	return g.sshKeys()
}

// sshKeys returns public keys based on
// provided private keys
func (g *Git) sshKeys() *gitssh.PublicKeys {

	if g.SSHKey == "" {
		log.Debug("no SSHKey provided")
		return nil
	}
	auth, err := gitssh.NewPublicKeys("git", []byte(g.SSHKey), "")
	if err != nil {
		log.Warningf("error on parsing private key: %s", err.Error())
		return nil
	}
	auth.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }
	return auth
}
