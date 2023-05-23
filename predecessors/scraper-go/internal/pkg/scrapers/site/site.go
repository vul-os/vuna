package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"scraper-go/internal/pkg/utils"

	"github.com/PuerkitoBio/goquery"
)

type SiteScraper struct {
	Client *http.Client
}

func New(
	client *http.Client,
) *SiteScraper {
	return &SiteScraper{
		Client: client,
	}
}

func (s *SiteScraper) ScrapeOne(url string) interface{} {
	currencyCode := "ZAR"

	name, image, technology := s.GetSiteInfo(url)
	siteID := utils.EncodeURL(url)

	currentDatetime := time.Now()
	formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

	fileName := fmt.Sprintf("site/%s_%s_site.csv", siteID, formattedDatetime)
	items := map[string]interface{}{
		"id":         siteID,
		"name":       strings.TrimSpace(name),
		"image":      strings.TrimSpace(image),
		"currency":   currencyCode,
		"technology": technology,
		"rate_limit": "1/s",
		"scraper":    technology,
	}
	fmt.Println(fileName, items)

	return items
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

