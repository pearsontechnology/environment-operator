package git

// Pull performs git pull for remote path
func (g *Git) Pull() error {
	tree, err := g.Repository.Worktree()

	if err != nil {
		return err
	}
	// @DEBUG: use spew for debug struct
	// spew.Dump(g.pullOptions())
	return tree.Pull(g.pullOptions())
}
