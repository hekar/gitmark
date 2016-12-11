package main

import (
	"os"
	"fmt"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	"gopkg.in/redis.v5"
	"github.com/libgit2/git2go"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

type Bookmark struct {
	Repo string
	Title string
	Url  string
}

type Error struct {
	Message string
	Status int
}

func commitBookmark(repo *git.Repository, b *Bookmark) (*Bookmark, error) {
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

	filename := "bookmarks/file.go"
	_, err = os.Create(filename)
	if err != nil {
		return nil, err
	}

	index, err := repo.Index()
	if err != nil {
		return nil, err
	}

	strategy := git.CheckoutForce
	err = repo.CheckoutIndex(index, &git.CheckoutOpts{
		Strategy: strategy,
	})
	if err != nil {
		return nil, err
	}

	err = index.AddByPath("file.go")
	if err != nil {
		return nil, err
	}

	treeOid, err := index.WriteTree()
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

	return b, nil
}

func saveBookmark(b *Bookmark) (*Bookmark, error) {
	client := redis.NewClient(&redis.Options{
		Addr: viper.GetString("RedisUrl"),
		Password: viper.GetString("RedisPassword"),
		DB: 0,
	})

	json, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	err = client.Set(b.Repo, json, 0).Err()
	if err != nil {
		return nil, err
	}

	return b, err
}

func addBookmark(c web.C, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered in f", rec)
			err, _ := rec.(error)
			apiError := &Error{
				Message: err.Error(),
				Status: 500,
			}

			encoder := json.NewEncoder(w)
			encoder.Encode(apiError)
		}
	}()

	repo := c.URLParams["repo"]
	title := r.PostForm.Get("title")
	url := r.PostForm.Get("url")
	bookmark := &Bookmark{
		Repo: repo,
		Title: title,
		Url: url,
	}

	_, err := saveBookmark(bookmark)
	if err != nil {
		panic(err)
	}

	origin := viper.GetString("RepoUrl")

	repository, err := git.OpenRepository(viper.GetString("Folder"))
	if err != nil {
		repository, err = git.Clone(
			origin, viper.GetString("Folder"), &git.CloneOptions{
				CheckoutBranch: viper.GetString("Branch"),
			})
		if err != nil {
			panic(err)
		}
	}

	defer repository.Free()

	_, err = commitBookmark(repository, bookmark)
	if err != nil {
		panic(err)
	}

	fmt.Print("d")
	encoder := json.NewEncoder(w)
	encoder.Encode(bookmark)
}

func setupViper() {
	viper.SetConfigName(".gitmrkrc")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}


func main() {
	setupViper()
	goji.Post("/bookmark/:repo", addBookmark)
	goji.Serve()
}
