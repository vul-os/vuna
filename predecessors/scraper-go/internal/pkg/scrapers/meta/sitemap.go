package meta

import (
	"fmt"
	"net/http"
	"regexp"
	"io/ioutil"
)

func ExtractSitemapURL(url string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", url+"/robots.txt", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for robots.txt: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request for robots.txt: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body for robots.txt: %s", err)
	}

	pattern := `(?is)sitemap:\s*([^\s]+)`
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(string(body))
	if len(match) > 1 {
		return match[1], nil
	}
	return "", nil
}
