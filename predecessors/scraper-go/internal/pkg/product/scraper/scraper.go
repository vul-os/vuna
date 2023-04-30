package products

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	products "scraper-go/internal/pkg/product"
	productStore "scraper-go/internal/pkg/product/store"
	site "scraper-go/internal/pkg/site"
	"scraper-go/utils"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
)

type api struct {
	store productStore.Store
}

func (a api) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", a.ScrapeOne) // POST /products/scrape - scrape a single url

	return r
}

// Robots.txt scraper
func (a api) ScrapeOne(w http.ResponseWriter, r *http.Request) {
	var s site.Site

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	baseUrl := strings.TrimSuffix(s.Url, "/")

	robotsTxtUrl := fmt.Sprintf("%s/robots.txt", baseUrl)
	log.Info().Msg(
		fmt.Sprintf(
			"robots.txt URL: %s",
			robotsTxtUrl,
		),
	)
	// Array containing all the known URLs
	knownUrls := []string{}
	// Create a Collector specifically for robots.txt and sitemap urls
	getUrlsCollector := colly.NewCollector()

	// Create a callback on the XPath query searching for the URLs
	getUrlsCollector.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {

		// Todo: sort this shit out
		if !(strings.Contains(e.Text, "/product/") ||
			strings.Contains(e.Text, "/products/") ||
			strings.Contains(e.Text, "bikemarket.co.za/shop/") ||
			strings.Contains(e.Text, "bottic.co.za/buy/")) {
			return
		}
		// old way, use it for logs
		knownUrls = append(knownUrls, e.Text)

		a.store.CreateOne(productStore.CreateOneRequest{
			Product: products.Product{
				Url:    e.Text, // e.Text is the product Url
				SiteId: s.Id,
			},
		})
	})

	getUrlsCollector.OnXML("//sitemapindex/sitemap/loc", func(e *colly.XMLElement) {
		if err := e.Request.Visit(e.Text); err != nil {
			log.Error().Err(err).Msg("unable to visit")
		}
	})
	// visit robots.txt and get the sitemaps
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Err getting hostname",
			),
		)
		panic(err)
	}
	hostName := u.Host
	robotsLines, err := utils.UrlToLines(robotsTxtUrl)
	if err != nil {
		log.Error().Err(err).Msg(
			"Err getting hostname",
		)
	}
	for _, line := range robotsLines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "sitemap:") {
			sitemapUrl := strings.Replace(lowerLine, "sitemap: ", "", 1)
			if !(strings.Contains(sitemapUrl, hostName)) {
				sitemapUrl = fmt.Sprintf("%s/%s", baseUrl, strings.TrimRight(sitemapUrl, "/"))
			}
			log.Info().Msg(
				fmt.Sprintf(
					"Sitemap Url: %s",
					sitemapUrl,
				),
			)
			err = getUrlsCollector.Visit(sitemapUrl)
			if err != nil {
				log.Error().Err(err).Msg(
					fmt.Sprintf(
						"Err Sitemap Url",
					),
				)
			}
		}
	}
	// Else visit the known sitemaps
	err = getUrlsCollector.Visit(fmt.Sprintf("%s/%s", baseUrl, "/sitemap_index.xml"))
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Error: sitemap_index.xml",
			),
		)
	}
	err = getUrlsCollector.Visit(fmt.Sprintf("%s/%s", baseUrl, "/sitemap.xml"))
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Error: sitemap.xml",
			),
		)
	}
	log.Err(err).Msg(
		fmt.Sprintf(
			"Error: sitemap.xml",
		),
	)

	log.Info().Msg(
		fmt.Sprintf(
			"Collected %d URLs",
			len(knownUrls),
		),
	)
	log.Info().Msg("Finished")
}
