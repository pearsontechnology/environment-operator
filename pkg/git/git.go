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
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// Git represents repository object and wraps git2go calls
type Git struct {
	SSHKey     string
	LocalPath  string
	RemotePath string
	BranchName string
	Repository *gogit.Repository
}

// Client is the git client
func Client() *Git {
	var repository *gogit.Repository
	var err error

	if _, err := os.Stat(config.Env.GitLocalPath); os.IsNotExist(err) {
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

	return &Git{
		LocalPath:  config.Env.GitLocalPath,
		RemotePath: config.Env.GitRepo,
		BranchName: config.Env.GitBranch,
		SSHKey:     config.Env.GitKey,
		Repository: repository,
	}
}

// EnvGitClient is git client for environment
func EnvGitClient(repo string, branch string, namespace string, env string) (*Git, error) {

	localPath := path.Join(config.Env.GitRootPath, namespace, env)

	var repository *gogit.Repository

	repository, err := gogit.PlainOpen(localPath)

	if err == gogit.ErrRepositoryNotExists {
		repository, err = gogit.PlainInit(localPath, false)
		if err != nil {
			return nil, fmt.Errorf("could not init local repository %s: %s", localPath, err.Error())
		}
	}

	remote, err := repository.Remote("origin")
	if err == gogit.ErrRemoteNotFound {
		_, err = repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{repo},
		})
		if err != nil {
			log.Errorf("could not attach to origin %s: %s", repo, err.Error())
		}
	}

	ref, err := repository.Head()

	// remote has been changed re-init repo
	if remote == nil || remote.Config().URLs[0] != repo || (err == nil && ref.Name().String() != branch) {
		os.RemoveAll(localPath)
		repository, err = gogit.PlainInit(localPath, false)
		if err != nil {
			return nil, fmt.Errorf("could not init local repository %s: %s", localPath, err.Error())
		}
		_, err = repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{repo},
		})
		if err != nil {
			log.Errorf("could not attach to origin %s: %s", repo, err.Error())
		}
	}

	return &Git{
		LocalPath:  localPath,
		RemotePath: remote.Config().URLs[0],
		BranchName: branch,
		SSHKey:     config.Env.GitKey,
		Repository: repository,
	}, nil
}

func (g *Git) pullOptions() *gogit.PullOptions {
	branch := fmt.Sprintf("refs/heads/%s", g.BranchName)
	return &gogit.PullOptions{
		ReferenceName: plumbing.ReferenceName(branch),
		Auth:          g.sshKeys(),
	}
}

func (g *Git) fetchOptions() *gogit.FetchOptions {
	return &gogit.FetchOptions{
		Auth: g.sshKeys(),
	}
}

func (g *Git) sshKeys() *gitssh.PublicKeys {
	if g.SSHKey == "" {
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
