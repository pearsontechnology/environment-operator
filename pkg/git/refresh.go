package git

import log "github.com/Sirupsen/logrus"

// Refresh checks if local git repository copy is outdated. If it is,
// changes are pulled in.
func (g *Git) Refresh() error {
	if ok, err := g.UpdatesExist(); ok {
		if err != nil {
			log.Error(err.Error())
			return err
		}
		log.Infof("Updates in repository: %s", g.RemotePath)
		g.CloneOrPull()
	}
	return nil
}
