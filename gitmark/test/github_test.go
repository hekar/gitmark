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
	client.Commit(bookmark)

	t.Log("Completed")
}

