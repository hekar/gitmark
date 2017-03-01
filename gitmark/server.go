package gitmark

import (
	"os"
	"fmt"
	"path"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strings"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/spf13/viper"
)

type Error struct {
	Message string
	Status int
}

func routeAddBookmark(c web.C, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered: ", rec)
			err, _ := rec.(error)
			apiError := &Error{
				Message: err.Error(),
				Status: 500,
			}

			encoder := json.NewEncoder(w)
			encoder.Encode(apiError)
		}
	}()

	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	repo := strings.Replace(c.URLParams["repo"], "%2F", "/", -1)
	title := r.PostForm.Get("title")
	url := r.PostForm.Get("url")
	bookmark := Bookmark{
		Repo: repo,
		Title: title,
		Url: url,
	}

	origin := viper.GetString("RepoUrl")
	branch := viper.GetString("Branch")
	rootFolder, err := ioutil.TempDir("", "gitmark-")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(rootFolder)

	repoFolder := path.Join(rootFolder, repo)
	fmt.Println(repoFolder)

	provider, err := CreateOrOpenRepository(repoFolder, origin, branch)
	defer provider.Free()

	root := RootFolder{
		Repo: origin,
		Path: rootFolder,
	}

	filename, err := AppendBookmark(root, bookmark)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))

	_, err = provider.commit(bookmark)
	if err != nil {
		panic(err)
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(bookmark)
}

func SetupViper() {
	viper.SetConfigName(".gitmarkrc")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func SetupServer() {
	SetupViper()
	goji.Post("/bookmark/:repo", routeAddBookmark)
	goji.Serve()
}

