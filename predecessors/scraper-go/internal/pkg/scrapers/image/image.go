package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

func UploadImageToGCS(imageURL string, siteID uuid.UUID, bucketName string) error {
	// Set up Google Cloud Storage client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Fetch the image from the URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	imageBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	gcsUrl := GetGcsUrl(imageURL, siteID)

	// Upload the image to Google Cloud Storage
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(gcsUrl)
	w := obj.NewWriter(ctx)
	_, err = w.Write(imageBytes)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func GetGcsUrl(imageURL string, siteID uuid.UUID) string {
	// Calculate the hash of the image URL
	urlHasher := md5.New()
	urlHasher.Write([]byte(imageURL))
	urlHash := hex.EncodeToString(urlHasher.Sum(nil))

	fileExt := filepath.Ext(imageURL)

	// Construct the object name with the hash and correct file extension
	return fmt.Sprintf("%s/%s.%s", siteID.String(), urlHash, fileExt)
}
