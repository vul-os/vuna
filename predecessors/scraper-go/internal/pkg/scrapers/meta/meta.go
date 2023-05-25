package meta

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	storage "github.com/imranparuk/scraper-go/internal/pkg/storage"
	"github.com/imranparuk/scraper-go/internal/pkg/utils"
)

type MetaScraper struct {
	Client      *http.Client
	FileStorage storage.FileStorage
}

func New(
	client *http.Client,
	fs storage.FileStorage,
) *MetaScraper {
	return &MetaScraper{
		Client:      client,
		FileStorage: fs,
	}
}

func (s *MetaScraper) ScrapeOne(url string) ([]string, error) {
	sitemapURL, err := s.extractSitemapURL(url)
	if err != nil {
		return nil, err
	}

	urls, err := s.extractURLsFromXML(sitemapURL, true)
	if err != nil {
		return nil, err
	}

	uniqueURLs := make(map[string]bool)

	for _, url := range urls {
		xmlURLs, err := s.extractURLsFromXML(url, true)
		if err != nil {
			return nil, err
		}
		for _, xmlURL := range xmlURLs {
			// fmt.Println(xmlURL, utils.SliceInString(xmlURL, []string{"/product/", "/products/"}))
			if utils.SliceInString(xmlURL, []string{"/product/", "/products/"}) {
				uniqueURLs[xmlURL] = true
			}
		}
	}

	result := make([]string, 0, len(uniqueURLs))
	for url := range uniqueURLs {
		result = append(result, url)
	}
	if s.FileStorage != nil {
		siteUrl := utils.RemoveURLPrefix(url)
		encodedSite := utils.EncodeURL(siteUrl)
		currentDatetime := time.Now()
		formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")
		fileName := fmt.Sprintf("meta/%s_%s_products.txt", encodedSite, formattedDatetime)
		err := s.FileStorage.WriteData(result, fileName)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *MetaScraper) extractSitemapURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url + "/robots.txt", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for robots.txt: %s", err)
	}

	resp, err := s.Client.Do(req)
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

func (s *MetaScraper) extractURLsFromXML(url string, allowRedirects bool) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %s", url, err)
	}

	client := s.Client
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
