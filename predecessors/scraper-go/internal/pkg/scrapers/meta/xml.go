package meta

import (
	"net/http"
	"fmt"
	"regexp"
	"io/ioutil"
)

func ExtractURLsFromXML(url string, client *http.Client, allowRedirects bool) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %s", url, err)
	}

	if !allowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request for %s: %s", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body for %s: %s", url, err)
	}

	urls := extractURLs(string(body))
	return urls, nil
}

func extractURLs(xmlData string) []string {
	pattern := `<loc>([^<]+)</loc>`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(xmlData, -1)
	urls := make([]string, len(matches))
	for i, match := range matches {
		urls[i] = match[1]
	}
	return urls
}
