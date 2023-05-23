package utils

import (
	"regexp"
	"strings"
	"net/url"
	"fmt"
)

func GetBaseURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return baseURL, nil
}

func RemoveURLPrefix(url string) string {
	pattern := `^(https?://)?(www\.)?`
	re := regexp.MustCompile(pattern)
	result := re.ReplaceAllString(url, "")
	return result
}

func SliceInString(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}