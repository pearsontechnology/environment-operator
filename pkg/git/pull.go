package git

import (
	gogit "gopkg.in/src-d/go-git.v4"
)

// Pull performs git pull for remote path
func (g *Git) Pull() error {
	tree, err := g.Repository.Worktree()

	if err != nil {
		return err
	}

	err = tree.Pull(g.pullOptions())

	if err != gogit.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}
