package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
    "scraper-go/internal/pkg/storage"
)

type FileStorageLocal struct {
	BaseDir string
}

func New(
	baseDir string,
) storage.FileStorage {
	return &FileStorageLocal{
		BaseDir: baseDir,
	}
}

func (s *FileStorageLocal) WriteData(data []byte, filePath string) error {
	fullPath := filepath.Join(s.BaseDir, filePath)

	// Ensure the directory for the file exists.
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(fullPath, data, 0644)
}

func (s *FileStorageLocal) ReadData(filePath string) ([]byte, error) {
	fullPath := filepath.Join(s.BaseDir, filePath)
	return ioutil.ReadFile(fullPath)
}

func (s *FileStorageLocal) GetLatestFiles(folderPrefix, textIn string) ([]string, error) {
	fullPath := filepath.Join(s.BaseDir, folderPrefix)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var latestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), textIn) {
			latestFiles = append(latestFiles, file.Name())
		}
	}

	return latestFiles, nil
}

func (s *FileStorageLocal) GetLatestFile(folderPrefix, textIn string) (string, error) {
	fullPath := filepath.Join(s.BaseDir, folderPrefix)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return "", err
	}

	var latestFile string
	var latestTime time.Time
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), textIn) {
			if file.ModTime().After(latestTime) {
				latestFile = file.Name()
				latestTime = file.ModTime()
			}
		}
	}

	if latestFile == "" {
		return "", errors.New("no file found")
	}

	return latestFile, nil
}

func (s *FileStorageLocal) CreateDirsForFile(filePath string) error {
	fullPath := filepath.Join(s.BaseDir, filePath)
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
