package products

import (
	"fmt"
	"net/http"
	"net/url"
	"scraper-go/utils"
	"strings"

	productsStore "scraper-go/internal/pkg/scrapers/products/store"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
)

type api struct {
	store productsStore.Store
}

// Routes creates a REST router for the products resource
func (a api) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", s.FindOne) // GET /products/{id} - read a single products by :id
		// r.Delete("/", a.Delete) // DELETE /products/{id} - delete a single products by :id
	})

	return r
}


// Robots.txt scraper
func (a api) FindOne(w http.ResponseWriter, r *http.Request) {

	baseUrl = strings.TrimSuffix(baseUrl, "/")

	
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
		//knownUrls = append(knownUrls, e.Text)
		if strings.Contains(e.Text, "/product/") ||
			strings.Contains(e.Text, "/products/") ||
			strings.Contains(e.Text, "bikemarket.co.za/shop/") ||
			strings.Contains(e.Text, "bottic.co.za/buy/") {
			knownUrls = append(knownUrls, e.Text)
		}
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
			fmt.Sprintf(
				"Err getting hostname",
			),
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
	fmt.Println("NumUrls", knownUrls)
	// for _, knownUrl := range knownUrls {
	// 	// log.Info().Msg(
	// 	// 	fmt.Sprintf(
	// 	// 		"Visiting Url: %s",
	// 	// 		knownUrl,
	// 	// 	),
	// 	// )
	// 	// if err := productsCollector.Visit(knownUrl); err != nil {
	// 	// 	log.Error().Err(err).Msg("unable to visit")
	// 	// }
	// }
	log.Info().Msg(
		fmt.Sprintf(
			"Collected %d URLs",
			len(knownUrls),
		),
	)
	log.Info().Msg("Finished")
}
