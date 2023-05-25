package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/imranparuk/scraper-go/internal/pkg/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/imranparuk/scraper-go/internal/pkg/storage"
)

// SiteData is a struct representing the site data
type SiteData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`

	Currency string `json:"currency"`

	Technology string `json:"technology"`
	Scraper    string `json:"scraper"`

	RateLimit string `json:"rate_limit"`
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

func (s *SiteScraper) ScrapeOne(url string) (map[string]interface{}, error) {
	currencyCode := "ZAR"

	name, image, technology := s.GetSiteInfo(url)
	siteID := utils.EncodeURL(url)

	items := map[string]interface{}{
		"id":         siteID,
		"name":       strings.TrimSpace(name),
		"image":      strings.TrimSpace(image),
		"currency":   currencyCode,
		"technology": technology,
		"rate_limit": "1/s",
		"scraper":    technology,
	}

	if s.FileStorage != nil {
		siteUrl := utils.RemoveURLPrefix(url)
		encodedSite := utils.EncodeURL(siteUrl)

		currentDatetime := time.Now()
		formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

		fileName := fmt.Sprintf("site/%s_%s_site.csv", encodedSite, formattedDatetime)
		err := s.FileStorage.WriteData([]map[string]interface{}{items}, fileName)
		if err != nil {
			return nil, err
		}
	}
	return items, nil
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
	name := document.Find("title").Text()

	// Extract the image of the website
	var image string
	document.Find("meta[property='og:image']").Each(func(i int, selection *goquery.Selection) {
		image, _ = selection.Attr("content")
	})

	technology := Detect(document)

	// Return the name, image, and technology of the website
	return name, image, technology
}
