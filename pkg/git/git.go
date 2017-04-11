package git

import (
	"os"

	"github.com/pearsontechnology/environment-operator/pkg/config"

	git2go "gopkg.in/libgit2/git2go.v24"
)

// Git represents repository object and wraps git2go calls
type Git struct {
	SSHKey     string
	LocalPath  string
	RemotePath string
	BranchName string
}

func Client() *Git {
	cfg := config.Load()

	return &Git{
		LocalPath:  cfg.GitLocalPath,
		RemotePath: cfg.GitRepo,
		BranchName: cfg.GitBranch,
		SSHKey:     cfg.GitKey,
	}
}

// CloneOrPull checks if repo exists in local path. If it does, it
// pulls changes from remotePath, if it doesn't, performs a full git clone
func (g *Git) CloneOrPull() error {
	if err := os.RemoveAll(g.LocalPath); err != nil {
		return err
	}
	return g.Clone()
}

func (g *Git) credentialsCallback(url string, username string, allowedTypes git2go.CredType) (git2go.ErrorCode, *git2go.Cred) {
	ret, cred := git2go.NewCredSshKeyFromMemory("git", "", g.SSHKey, "")
	return git2go.ErrorCode(ret), &cred
}

// Made this one just return 0 during troubleshooting...
func certificateCheckCallback(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
	return 0
}

func (g *Git) cloneOptions() *git2go.CloneOptions {
	opts := &git2go.CloneOptions{}
	opts.FetchOptions = g.fetchOptions()
	return opts
}

func (g *Git) fetchOptions() *git2go.FetchOptions {
	return &git2go.FetchOptions{
		RemoteCallbacks: git2go.RemoteCallbacks{
			CredentialsCallback:      g.credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		},
	}
}
