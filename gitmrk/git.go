package gitmrk

import (
	"fmt"
	"time"
	"github.com/libgit2/git2go"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

func CreateOrOpenRepository(folder string, origin string, branch string) (*git.Repository, error) {
	fmt.Println(folder)
	repository, err := git.OpenRepository(folder)
	if err != nil {
		repository, err = git.Clone(
			origin, folder, &git.CloneOptions{
				CheckoutBranch: branch,
			})
		if err != nil {
			return repository, err
		}
	}

	return repository, nil
}

func commitBookmark(repo *git.Repository, b Bookmark) (*Bookmark, error) {
	refname := "refs/heads/" + viper.GetString("Branch")
	committer := &git.Signature{
		Name: viper.GetString("UserName"),
		Email: viper.GetString("UserEmail"),
		When: time.Now(),
	}
	author := committer

	message := viper.GetString("MessagePrefix") + b.Title

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	index, err := repo.Index()
	if err != nil {
		return nil, err
	}

	paths := []string {"README.md"}
	err = index.AddAll(paths, git.IndexAddForce, nil)
	if err != nil {
		return nil, err
	}

	treeOid, err := index.WriteTreeTo(repo)
	if err != nil {
		return nil, err
	}

	oid := head.Target()

	parent, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	tree, err := repo.LookupTree(treeOid)
	if err != nil {
		return nil, err
	}

	_, err = repo.CreateCommit(refname,
		author, committer, message, tree, parent)
	if err != nil {
		return nil, err
	}

	refspecs := []string {refname}
	spew.Dump(refspecs)

	callbacks := git.RemoteCallbacks {
	}
	options := &git.PushOptions{
		RemoteCallbacks: callbacks,
		PbParallelism: 0,
	}

	remote, err := repo.Remotes.Lookup(viper.GetString("Remote"))
	if err != nil {
		return nil, err
	}

	err = remote.Push(refspecs, options)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

