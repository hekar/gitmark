package gitmrk

import (
	"os"
	"path"
	"testing"
	"io/ioutil"
	"github.com/gitmrk/gitmrk"
)

// Integration test
// Test the creating of a bookmark
func TestAppendBookmark(t *testing.T) {
	repo := "hekar/bookmarks"
	url := "http://google.ca"
	title := "Google"

	rootPath, err := ioutil.TempDir("", "gitmrk-")
	if err != nil {
		t.Error(err)
	}

	repoPath := path.Join(rootPath, repo)
	err = os.MkdirAll(repoPath, 0755)
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(repoPath)

	t.Log("Repo path:", repoPath)

	_, err = os.OpenFile(path.Join(repoPath, "README.md"), os.O_CREATE, 0755)
	if err != nil {
		t.Error(err)
	}

	root := gitmrk.RootFolder{
		Repo: repo,
		Path: rootPath,
	}
	bookmark := gitmrk.Bookmark{
		Repo: repo,
		Url: url,
		Title: title,
	}
	gitmrk.AppendBookmark(root, bookmark)

	file, err := os.OpenFile(path.Join(repoPath, "README.md"), os.O_RDONLY, 0755)
	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	t.Log("Contents:", string(content))
}

