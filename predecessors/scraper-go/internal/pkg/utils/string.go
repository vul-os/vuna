package utils

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
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

// RandomSample returns a random sample of strings from a given list
func RandomSample(stringList []string, sampleSize int) []string {
	n := len(stringList)
	if sampleSize >= n {
		// Return the entire list if the sample size is equal to or larger than the list size
		return stringList
	}

	// Create a slice to store the sampled strings
	sampledStrings := make([]string, sampleSize)

	// Generate random indices without replacement
	for i := 0; i < sampleSize; i++ {
		// Generate a random index between 0 and n-i-1
		randomIndex := rand.Intn(n - i)

		// Move the selected element to the end of the list
		stringList[n-i-1], stringList[randomIndex] = stringList[randomIndex], stringList[n-i-1]

		// Add the selected string to the sampled list
		sampledStrings[i] = stringList[n-i-1]
	}

	return sampledStrings
}

// CleanString removes leading and trailing white spaces,
// newline characters, tab characters, spaces, single and double quotes from the input string
func CleanString(input string) string {
	cleanStr := strings.ReplaceAll(input, "\n", "")
	cleanStr = strings.ReplaceAll(cleanStr, "\t", "")
	cleanStr = strings.TrimSpace(cleanStr)
	cleanStr = strings.ReplaceAll(cleanStr, "\"", "")
	cleanStr = strings.ReplaceAll(cleanStr, "'", "")
	cleanStr = strings.ReplaceAll(cleanStr, "\"", "")

	return cleanStr
}

func CleanStringCSV(value string) string {
	value = CleanString(value)
	if len(strings.TrimSpace(value)) == 0 {
		value = "\\N"
	}
	return value
}
