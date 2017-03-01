package gitmark

import (
	"time"
	"github.com/libgit2/git2go"
	"github.com/spf13/viper"
)

type GitProvider struct {
	Repository *git.Repository
	Folder string
	Origin string
	Branch string
}

func CreateOrOpenRepository(folder string, origin string, branch string) (*GitProvider, error) {
	repository, err := git.OpenRepository(folder)
	if err != nil {
		repository, err = git.Clone(
			origin, folder, &git.CloneOptions{
				CheckoutBranch: branch,
			})
		if err != nil {
			return nil, err
		}
	}

	provider := new(GitProvider)
	provider.Repository = repository
	provider.Folder = folder
	provider.Origin = origin
	provider.Branch = branch
	return provider, nil
}

func (g *GitProvider) Free() {
	g.Repository.Free()
}

func (g *GitProvider) commit(b Bookmark) (*Bookmark, error) {
	repo := g.Repository

	username := viper.GetString("UserName")
	email := viper.GetString("UserEmail")
	message := viper.GetString("MessagePrefix") + b.Title
	remoteName := viper.GetString("Remote")
	refname := "refs/heads/" + g.Branch

	committer := &git.Signature{
		Name: username,
		Email: email,
		When: time.Now(),
	}

	author := committer

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

	callbacks := git.RemoteCallbacks {
	}
	options := &git.PushOptions{
		RemoteCallbacks: callbacks,
		PbParallelism: 0,
	}

	remote, err := repo.Remotes.Lookup(remoteName)
	if err != nil {
		return nil, err
	}

	refspecs := []string {refname}
	err = remote.Push(refspecs, options)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

