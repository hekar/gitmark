package gitmark

import (
	"time"
	"context"
	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"github.com/davecgh/go-spew/spew"
)

type GithubProvider struct {
	Context context.Context
	Client *github.Client
	Repository *github.Repository
	Owner string
	Repo string
}

func NewGithubClient(owner string, repo string) (*GithubProvider, error) {

	accessToken := viper.GetString("github_access_token")
	spew.Println("Connecting with ", accessToken)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repository, response, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		spew.Println(response)
		return nil, err
	}

	provider := new(GithubProvider)
	provider.Context = ctx
	provider.Client = client
	provider.Repository = repository
	provider.Owner = owner
	provider.Repo = repo
	return provider, nil
}

func (g *GithubProvider) Free() {
	/* intentionally blank */
}

func (g *GithubProvider) Commit(b Bookmark) (*Bookmark, error) {
	username := viper.GetString("UserName")
	email := viper.GetString("UserEmail")
	message := viper.GetString("MessagePrefix") + b.Title
	refName := "heads/" + viper.GetString("Branch")

	now := time.Now()
	committer := &github.CommitAuthor {
		Date: &now,
		Name: &username,
		Email: &email,
	}

	git := g.Client.Git
	existingReference, _, err := git.GetRef(g.Context, g.Owner, g.Repo, refName)
	if err != nil {
		return nil, err
	}

	spew.Println(existingReference)

	tree, _, err := git.GetTree(g.Context, g.Owner, g.Repo, existingReference.Object.GetSHA(), false)
	if err != nil {
		return nil, err
	}

	spew.Println("tree", tree)
	head, _, err := git.GetCommit(g.Context, g.Owner, g.Repo, tree.GetSHA())
	if err != nil {
		return nil, err
	}

	parents := []github.Commit{*head}

	spew.Println("head", head)
	commit := &github.Commit{
		Author: committer,
		Committer: committer,
		Message: &message,
		Tree: head.Tree,
		Parents: parents,
	}

	createdCommit, _, err := g.Client.Git.CreateCommit(g.Context, g.Owner, g.Repo, commit)
	if err != nil {
		return nil, err
	}
	spew.Println(createdCommit)

	sha := createdCommit.GetSHA()
	url := createdCommit.GetURL()
	object := &github.GitObject{
		SHA: &sha,
		URL: &url,
	}

	reference := &github.Reference{
		Ref: &refName,
		Object: object,
	}

	ref, _, err := git.UpdateRef(g.Context, g.Owner, g.Repo, reference, true)
	if err != nil {
		return nil, err
	}
	spew.Println(ref)
	spew.Println(createdCommit)


	return &b, nil
}

