package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
)

// SiteData is a struct representing the site data
type SiteData struct {
	SiteIdentifier string `json:"site_identifier"`

	Name  string `json:"name"`
	Image string `json:"image"`

	Currency string `json:"currency"`

	Technology string `json:"technology"`
	Scraper    string `json:"scraper"`

	RateLimit string `json:"rate_limit"`

	Url string `json:"url"`
}

// func StructToMap(s SiteData) map[string]string {
// 	return map[string]string{
// 		"id":         s.ID,
// 		"name":       s.Name,
// 		"image":      s.Image,
// 		"currency":   s.Currency,
// 		"technology": s.Technology,
// 		"scraper":    s.Scraper,
// 		"rate_limit": s.RateLimit,
// 	}
// }

type SiteScraper struct {
	Client      *http.Client
	FileStorage storage.FileStorage
}

func New(
	client *http.Client,
	fs storage.FileStorage,
) *SiteScraper {
	return &SiteScraper{
		Client:      client,
		FileStorage: fs,
	}
}

func (s *SiteScraper) ScrapeOne(url string) (map[string]string, error) {
	currencyCode := "ZAR"

	name, image, technology := s.GetSiteInfo(url)
	hostName, err := utils.GetHostName(url)
	if err != nil {
		return nil, err
	}
	hostIdentifier, _, err := utils.StringToIdentifier(url, nil)
	if err != nil {
		return nil, err
	}
	items := []map[string]string{{
		"site_identifier": hostIdentifier,
		"name":            strings.TrimSpace(name),
		"image":           strings.TrimSpace(image),
		"currency":        currencyCode,
		"technology":      technology,
		"rate_limit":      "1/s",
		"scraper":         technology,
		"url":             hostName,
	}}

	if s.FileStorage != nil {

		currentDatetime := time.Now()
		formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

		fileName := fmt.Sprintf("site/%s_%s_site.csv", hostName, formattedDatetime)
		err := s.FileStorage.WriteData(items, fileName)
		if err != nil {
			return nil, err
		}
	}
	return items[0], nil
}

func (s *SiteScraper) GetSiteInfo(url string) (string, string, string) {
	// Send an HTTP GET request to the website and retrieve the HTML content
	response, err := s.Client.Get(url)
	if err != nil {
		// Handle the error
	}

	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// Handle the error
	}

	// Parse the HTML content using goquery
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		// Handle the error
	}

	// Extract the name of the website
	name := document.Find("site_name").Text()

	// Extract the image URL from the og:image meta tag
	var image string
	document.Find("meta[property='og:image']").Each(func(i int, selection *goquery.Selection) {
		image, _ = selection.Attr("content")
	})

	// If og:image not found, try finding the first image in div#logo img
	if len(image) == 0 {
		image, _ = document.Find("div#logo img").First().Attr("src")
	}

	// If image is still not found, try finding the first image with link rel="icon"
	if len(image) == 0 {
		document.Find("link[rel='icon']").Each(func(i int, selection *goquery.Selection) {
			iconURL, exists := selection.Attr("href")
			if exists && strings.HasSuffix(iconURL, ".png") || strings.HasSuffix(iconURL, ".jpg") {
				image = iconURL
				return
			}
		})
	}
	technology := Detect(response, body)
	// Return the name, image, and technology of the website
	return name, image, technology
}
