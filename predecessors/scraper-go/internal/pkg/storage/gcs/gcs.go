package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"

	"time"

	"google.golang.org/api/iterator"

	cloudStorage "cloud.google.com/go/storage"
)

type FileStorageGCS struct {
	BucketName string
	Client     cloudStorage.Client
}

func New(
	bucketName string,
	client cloudStorage.Client,
) storage.FileStorage {
	return &FileStorageGCS{
		BucketName: bucketName,
		Client:     client,
	}
}

func (s *FileStorageGCS) WriteData(data interface{}, filePath string) error {
	ctx := context.Background()
	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(filePath)
	writer := obj.NewWriter(ctx)

	switch d := data.(type) {
	case []string:
		err := WriteTextFile(writer, d)
		if err != nil {
			return fmt.Errorf("error writing text file: %v", err)
		}
	case []map[string]string:
		err := WriteCSVFile(writer, d)
		if err != nil {
			return fmt.Errorf("error writing CSV file: %v", err)
		}
	default:
		return fmt.Errorf("unsupported data type")
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing writer to GCS: %v", err)
	}

	return nil
}

func (s *FileStorageGCS) ReadData(filePath string) (interface{}, error) {
	ctx := context.Background()
	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(filePath)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening reader for file: %v", err)
	}
	defer reader.Close()

	extension := GetFileExtension(filePath)

	if extension == ".txt" {
		data, err := ReadTextFile(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading text file: %v", err)
		}
		return data, nil
	} else if extension == ".csv" {
		data, err := ReadCSVFile(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading CSV file: %v", err)
		}
		return data, nil
	} else {
		return nil, fmt.Errorf("unsupported file format: %s", extension)
	}
}

func (s *FileStorageGCS) GetLatestFiles(folderPrefix, textIn string) ([]string, error) {
	ctx := context.Background()
	bucket := s.Client.Bucket(s.BucketName)

	query := &cloudStorage.Query{
		Delimiter: "/",
		Prefix:    folderPrefix,
	}

	latestFilesMap := make(map[string]string)

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error getting next object from GCS: %v", err)
		}

		if attrs.Prefix == "" && strings.Contains(attrs.Name, textIn) {
			splitName := strings.Split(attrs.Name, "_")
			if len(splitName) < 2 {
				continue
			}
			url := splitName[0]
			dateAndTime := strings.Split(splitName[1], "-")
			if len(dateAndTime) < 3 {
				continue
			}

			dateString := strings.Join(dateAndTime[:3], "-")
			newTime, err := time.Parse("2006-01-02", dateString)
			if err != nil {
				return nil, fmt.Errorf("error parsing date: %v", err)
			}

			existingFile, exists := latestFilesMap[url]
			if exists {
				existingDateString := strings.Split(existingFile, "_")[1]
				existingDate, _ := time.Parse("2006-01-02", existingDateString[:10])
				if newTime.After(existingDate) {
					latestFilesMap[url] = attrs.Name
				}
			} else {
				latestFilesMap[url] = attrs.Name
			}
		}
	}

	var latestFiles []string
	for _, filename := range latestFilesMap {
		latestFiles = append(latestFiles, filename)
	}

	return latestFiles, nil
}

func (s *FileStorageGCS) GetLatestFile(folderPrefix, textIn string) (string, error) {
	ctx := context.Background()
	bucket := s.Client.Bucket(s.BucketName)

	query := &cloudStorage.Query{
		Delimiter: "/",
		Prefix:    folderPrefix,
	}

	var latestFile string
	var latestTime time.Time

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error getting next object from GCS: %v", err)
		}
		if attrs.Prefix == "" && strings.Contains(attrs.Name, textIn) {
			newName := strings.ReplaceAll(attrs.Name, textIn, "")
			splitName := strings.Split(newName, "_")
			dateString := ""
			if len(splitName) == 2 {
				// format: [date]_[textIn]
				dateString = splitName[0]
			} else if len(splitName) == 3 {
				// format: [url]_[date]_[textIn]
				dateString = splitName[1]
			} else {
				continue
			}

			dateAndTime := strings.Split(dateString, "-")
			if len(dateAndTime) < 3 {
				continue
			}

			date := strings.Join(dateAndTime[:3], "-")
			newTime, err := time.Parse("2006-01-02", date)
			if err != nil {
				return "", fmt.Errorf("error parsing date: %v", err)
			}

			if newTime.After(latestTime) {
				latestFile = newName
				latestTime = newTime
			}
		}
	}

	return latestFile, nil
}
