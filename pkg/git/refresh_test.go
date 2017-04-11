package git

import "testing"

func TestRefreshError(t *testing.T) {
	remotePath := createTestRepo(t)
	localPath := createSrcPath(t)
	defer cleanupTestPath(localPath)
	defer cleanupTestPath(remotePath)

	g := initAndClone(t, localPath, remotePath)
	g.LocalPath = "/nonexistent"

	ret := g.Refresh()
	if ret == nil {
		t.Errorf("Expected failure, got: %s", ret.Error())
	}
}

func TestRefreshSuccess(t *testing.T) {
	remotePath := createTestRepo(t)
	localPath := createSrcPath(t)
	defer cleanupTestPath(localPath)
	defer cleanupTestPath(remotePath)

	g := initAndClone(t, localPath, remotePath)
	commitTestJunk(t, remotePath, "zzz.bitesize")

	ret := g.Refresh()
	if ret != nil {
		t.Errorf("Expected success, got: %s", ret.Error())
	}

}
