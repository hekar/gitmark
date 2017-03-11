package gitmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type RootFolder struct {
	Repo string
	Path string
}

type Bookmark struct {
	Repo  string
	Title string
	Url   string
}

func appendBookmark(root RootFolder, bookmark Bookmark) (string, error) {
	appendJson := false
	appendContent := []byte{0}
	if appendJson {
		var err error
		appendContent, err = json.Marshal(bookmark)
		if err != nil {
			return "", err
		}
	} else {
		appendContent = []byte(fmt.Sprintf("\n* [%s](%s)", bookmark.Title, bookmark.Url))
	}

	folder := path.Join(root.Path, bookmark.Repo)
	_, err := createMissingFolder(folder)
	if err != nil {
		return "", err
	}

	filename := path.Join(folder, "README.md")
	err = appendToFile(filename, appendContent)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func createMissingFolder(folder string) (bool, error) {
	folderExists, err := exists(folder)
	if err != nil {
		return false, err
	}

	if !folderExists {
		err = os.MkdirAll(folder, 0600)
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func appendToFile(filename string, content []byte) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteAt(content, 0); err != nil {
		return err
	}

	return nil
}
