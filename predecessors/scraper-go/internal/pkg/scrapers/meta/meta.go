package meta

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	storage "github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
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
	sitemapURL, err := ExtractSitemapURL(url, s.Client)
	if err != nil {
		fmt.Println("no sitemap in robots.txt")
	}

	sitemapUrls := []string{
		fmt.Sprintf("%s/sitemap_index.xml", url),
		fmt.Sprintf("%s/sitemap.xml", url),
		fmt.Sprintf("%s/wp_sitemap.xml", url),
		sitemapURL,
	}

	var urls []string
	for _, sitemapUrl := range sitemapUrls {
		tUrls, err := ExtractURLsFromXML(sitemapUrl, s.Client, true)
		if err != nil {
			fmt.Println("sitemap error: ", sitemapUrl, err)
			continue
		}
		if len(tUrls) > 0 {
			urls = append(urls, tUrls...)
		}

	}
	if len(urls) == 0 {
		return nil, errors.New("no urls found in all sitemaps")
	}
	uniqueURLs := make(map[string]bool)

	for _, url := range urls {
		xmlURLs, err := ExtractURLsFromXML(url, s.Client, true)
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
		currentDatetime := time.Now()
		formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

		hostName, _, err := utils.GetHostName(url)
		if err != nil {
			return nil, err
		}

		fileName := fmt.Sprintf("meta/%s_%s_products.txt", hostName, formattedDatetime)
		err = s.FileStorage.WriteData(result, fileName)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
