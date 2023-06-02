package bi

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func ProcessFile(name string) string {
	directory := filepath.Join(os.Getenv("PWD"), "sql")
	extension := ".sql"

	filename := name + extension
	filepath := filepath.Join(directory, filename)

	if fileExists(filepath) {
		fileContents, err := ioutil.ReadFile(filepath)
		if err != nil {
			// Handle error reading file
			return ""
		}
		return string(fileContents)
	} else {
		return ""
	}
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
