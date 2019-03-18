package git

import (
	"fmt"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"golang.org/x/crypto/ssh"
	gogit "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	gogithttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// Git represents repository object and wraps git2go calls
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

	return &Git{
		LocalPath:  config.Env.GitLocalPath,
		RemotePath: config.Env.GitRepo,
		BranchName: config.Env.GitBranch,
		SSHKey:     config.Env.GitKey,
		GitToken:   config.Env.GitToken,
		GitUser:    config.Env.GitUser,
		Repository: repository,
	}
}

func (g *Git) pullOptions() *gogit.PullOptions {
	branch := fmt.Sprintf("refs/heads/%s", g.BranchName)
	// Return options with token auth if enabled
	if g.GitToken != "" {
		return &gogit.PullOptions{
			ReferenceName: plumbing.ReferenceName(branch),
			Auth: &gogithttp.BasicAuth{
				Username: g.GitUser,
				Password: g.GitToken,
			},
		}
	}

	return &gogit.PullOptions{
		ReferenceName: plumbing.ReferenceName(branch),
		Auth:          g.sshKeys(),
	}
}

func (g *Git) fetchOptions() *gogit.FetchOptions {
	// Return options with token auth if enabled
	if g.GitToken != "" {
		return &gogit.FetchOptions{
			Auth: &gogithttp.BasicAuth{
				Username: g.GitUser,
				Password: g.GitToken,
			},
		}
	}
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
