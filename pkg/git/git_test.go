package git

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"

	gogit "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

func initAndClone(t *testing.T, local, remote string) *Git {
	repository, _ := gogit.PlainInit(local, false)
	repository.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{remote},
	})
	g := &Git{
		LocalPath:  local,
		RemotePath: remote,
		BranchName: "master",
		Repository: repository,
	}

	if err := g.Pull(); err != nil {
		t.Errorf("Error on initial pull: %s", err.Error())
	}
	return g
}

func initAndCloneToken(t *testing.T, local, remote string, gittoken string, gituser string) *Git {
	repository, _ := gogit.PlainInit(local, false)
	repository.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{remote},
	})
	g := &Git{
		LocalPath:  local,
		RemotePath: remote,
		BranchName: "master",
		Repository: repository,
		GitToken:   gittoken,
		GitUser:    gituser,
	}

	if err := g.Pull(); err != nil {
		t.Errorf("Error on initial pull: %s", err.Error())
	}
	return g
}

func createTestRepo(t *testing.T) string {
	path, err := ioutil.TempDir("", "env-operator")
	checkFatal(t, err, "temp dir create")

	barepath := path + "/bare"
	localpath := path + "/local"
	defer cleanupTestPath(localpath)

	_, err = gogit.PlainInit(barepath, true)
	checkFatal(t, err, "barepath init")

	localrepo, err := gogit.PlainInit(localpath, false)
	checkFatal(t, err, "localrepo init")

	w, err := localrepo.Worktree()
	checkFatal(t, err, "localrepo worktree")

	_, err = localrepo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{barepath},
	})
	checkFatal(t, err, "localrepo createremote")

	tmpfile := "environments.bitesize"
	err = ioutil.WriteFile(localpath+"/"+tmpfile, []byte("foo\n"), 0644)
	checkFatal(t, err)

	w.Add(tmpfile)

	_, err = w.Commit("initial commit", &gogit.CommitOptions{
		Author: &gitobject.Signature{
			Name:  "Author",
			Email: "author@pearson.com",
			When:  time.Now(),
		},
	})
	checkFatal(t, err)

	err = w.Checkout(&gogit.CheckoutOptions{Create: false, Branch: plumbing.ReferenceName("refs/heads/master")})
	checkFatal(t, err, "localrepo create branch")

	err = localrepo.Push(&gogit.PushOptions{RemoteName: "origin"})
	checkFatal(t, err)

	return barepath
}

func commitTestJunk(t *testing.T, dest string, file string) {
	var path string

	tempPath, err := ioutil.TempDir("", "env-operator")
	defer cleanupTestPath(tempPath)

	checkFatal(t, err, "commitTestJunk temp dir create")

	repo, err := gogit.PlainClone(tempPath, false, &gogit.CloneOptions{URL: dest})
	checkFatal(t, err)

	w, err := repo.Worktree()
	checkFatal(t, err)

	if file == "" {
		f, e := ioutil.TempFile(tempPath, "junk")
		path = f.Name()
		file = filepath.Base(path)
		checkFatal(t, e)
	} else {
		path = tempPath + "/" + file
	}

	contents := randomString(64)

	err = ioutil.WriteFile(path, contents, 0644)
	checkFatal(t, err)

	w.Add(file)
	_, err = w.Commit("test", &gogit.CommitOptions{
		Author: &gitobject.Signature{
			Name:  "Author",
			Email: "author@pearson.com",
			When:  time.Now(),
		},
	})
	repo.Push(&gogit.PushOptions{})
}

func checkFatal(t *testing.T, args ...interface{}) {
	var str string
	var err error

	if len(args) == 0 {
		return
	}
	if len(args) == 2 {
		str, _ = args[1].(string)
	}
	err, _ = args[0].(error)
	if err != nil {
		t.Fatalf("%s: %s", str, err.Error())
	}
}

func createSrcPath(t *testing.T) string {
	path, err := ioutil.TempDir("", "env-operator-local")
	checkFatal(t, err)
	return path
}

func cleanupTestPath(path string) {
	os.RemoveAll(path + "/")
}

func randomString(strlen int) []byte {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return result
}
