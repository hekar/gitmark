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

func (g *GithubProvider) getHead(ref string) (*github.Commit, error) {
	git := g.Client.Git
	existingReference, _, err := git.GetRef(g.Context, g.Owner, g.Repo, ref)
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

	return head, nil
}

func (g *GithubProvider) createCommit(message, username, email string, head *github.Commit) (*github.Commit, error) {
	parents := []github.Commit{*head}

	now := time.Now()
	committer := &github.CommitAuthor {
		Date: &now,
		Name: &username,
		Email: &email,
	}

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

	return createdCommit, nil
}

func (g *GithubProvider) updateReference(object *github.GitObject, ref string) (*github.Reference, error) {
	reference := &github.Reference{
		Ref: &ref,
		Object: object,
	}

	createdReference, _, err := g.Client.Git.UpdateRef(g.Context, g.Owner, g.Repo, reference, true)
	if err != nil {
		return nil, err
	}

	return createdReference, nil
}

func (g *GithubProvider) getReadme(ref string) (*github.RepositoryContent, error) {
	options := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	readme, _, err := g.Client.Repositories.GetReadme(g.Context, g.Owner, g.Repo, options)
	if err != nil {
		return nil, err
	}

	return readme, nil
}

func (g *GithubProvider) updateReadme(sha, message, branch string, content []byte) (*github.RepositoryContentResponse, error) {
	file := "README.md"
	options := &github.RepositoryContentFileOptions{
		SHA: &sha,
		Message: &message,
		Branch: &branch,
		Content: content,
	}

	readme, _, err := g.Client.Repositories.UpdateFile(g.Context, g.Owner, g.Repo, file, options)
	if err != nil {
		return nil, err
	}

	return readme, nil
}

func (g *GithubProvider) Commit(b Bookmark) (*Bookmark, error) {
	message := viper.GetString("MessagePrefix") + b.Title
	branch := viper.GetString("Branch")
	ref := "heads/" + branch

	head, err := g.getHead(ref)
	if err != nil {
		return nil, err
	}

	spew.Println("head", head)

	readme, err := g.getReadme(ref)
	if err != nil {
		return nil, err
	}

	readmeContents, err := readme.GetContent()
	if err != nil {
		return nil, err
	}

	spew.Println("readme", readmeContents)

	updatedReadme, err := g.updateReadme(readme.GetSHA(), message, branch, []byte(readmeContents + "\n* [" + b.Title + "](" + b.Url + ")"))
	if err != nil {
		return nil, err
	}

	spew.Println("updatedReadme", updatedReadme)

	return &b, nil
}

