package meta

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func SliceInString(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

func MetaScrapeOne(url string) []string {
	// Extract sitemap URL from robots.txt
	sitemapURL, err := extractSitemapURL(url)
	if err != nil {
		fmt.Println("Error extracting sitemap URL:", err)
		return nil
	}

	// Extract URLs from the sitemap
	urls, err := extractURLsFromXML(sitemapURL, true)
	if err != nil {
		fmt.Println("Error extracting URLs from sitemap:", err)
		return nil
	}
	// Store the unique URLs
	uniqueURLs := make(map[string]bool)

	// Process each URL
	for _, url := range urls {
		// Extract URLs from XML files
		xmlURLs, err := extractURLsFromXML(url, true)
		if err != nil {
			fmt.Println("Error extracting URLs from", url, ":", err)
			continue
		}

		// Add extracted URLs to the unique URLs map
		for _, xmlURL := range xmlURLs {
			fmt.Println(xmlURL, SliceInString(xmlURL, []string{"/products/", "/products/"}))
			if SliceInString(xmlURL, []string{"/product/", "/products/"}) {
				uniqueURLs[xmlURL] = true
			}
		}
	}

	// Convert unique URLs to a slice of strings
	result := make([]string, 0, len(uniqueURLs))
	for url := range uniqueURLs {
		result = append(result, url)
	}

	return result
}

func extractSitemapURL(url string) (string, error) {
	// Create an HTTP client
	client := &http.Client{}

	// Create a new GET request for robots.txt
	req, err := http.NewRequest("GET", url+"/robots.txt", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for robots.txt: %s", err)
	}

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request for robots.txt: %s", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body for robots.txt: %s", err)
	}
	// Find the sitemap URL in robots.txt
	pattern := `(?is)sitemap:\s*([^\s]+)`
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(string(body))
	if len(match) > 1 {
		return match[1], nil
	}
	return "", nil
}

func extractURLsFromXML(url string, allowRedirects bool) ([]string, error) {
	// Create an HTTP client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !allowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %s", url, err)
	}

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request for %s: %s", url, err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body for %s: %s", url, err)
	}
	// Extract URLs from the XML response
	urls := extractURLs(string(body))
	return urls, nil
}

func extractURLs(xmlData string) []string {
	// Define the regular expression pattern to match URLs
	pattern := `<loc>([^<]+)</loc>`

	// Compile the regex pattern
	regex := regexp.MustCompile(pattern)

	// Find all matches in the XML data
	matches := regex.FindAllStringSubmatch(xmlData, -1)
	// Extract and return the URLs
	urls := make([]string, len(matches))
	for i, match := range matches {
		urls[i] = match[1]
	}

	return urls
}
