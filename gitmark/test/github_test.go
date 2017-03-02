package gitmark

import (
	"testing"
	"github.com/hekar/gitmark"
)

// Integration test
// Test the creating of a bookmark
func TestCreateGithubClient(t *testing.T) {
	gitmark.SetupViper()

	owner := "hekar"
	repo := "bookmarks"
	client, err := gitmark.NewGithubClient(owner, repo)
	if err != nil {
		t.Fatal("Err:", err)
	}

	bookmark := gitmark.Bookmark {
	}

	_, err = client.Commit(bookmark)
	if err != nil {
		t.Fatal("Err:", err)
	}

	t.Log("Completed")
}

