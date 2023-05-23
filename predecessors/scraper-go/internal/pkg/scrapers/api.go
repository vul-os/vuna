package scrapers

import (
	"net/http"
	"net/url"

	meta "scraper-go/internal/pkg/scrapers/meta"
	site "scraper-go/internal/pkg/scrapers/site"
	// product "scraper-go/internal/pkg/scrapers/product"

	"scraper-go/internal/pkg/storage"

	"github.com/gin-gonic/gin"
)

type ScraperAPI struct {
	FileStorage storage.FileStorage
	Proxies     []string
}

func New(
	fs storage.FileStorage,
	p []string,
) *ScraperAPI {
	return &ScraperAPI{
		FileStorage: fs,
		Proxies:     p,
	}
}

func (api *ScraperAPI) Meta(c *gin.Context, baseURL string) {
	baseURL, err := url.QueryUnescape(baseURL)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	client := http.Client{}
	scraper := meta.New(&client, api.FileStorage)
	metaData := scraper.ScrapeOne(baseURL)

	c.JSON(http.StatusOK, metaData)
}

func (api *ScraperAPI) Site(c *gin.Context, baseURL string) {
	baseURL, err := url.QueryUnescape(baseURL)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	client := http.Client{}
	scraper := site.New(&client, api.FileStorage)
	siteData := scraper.ScrapeOne(baseURL)

	c.JSON(http.StatusOK, siteData)
}

// func (api *ScraperAPI) Product(c *gin.Context, productURL string) {
// 	productURL, err := url.QueryUnescape(productURL)
// 	if err != nil {
// 		c.String(http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	var proxies []string
// 	if err := c.ShouldBindJSON(&proxies); err != nil {
// 		c.String(http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	var scraperCode string
// 	if err := c.ShouldBindJSON(&scraperCode); err != nil {
// 		c.String(http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	client := http.Client{}
// 	scraper := product.New(&client, api.FileStorage)
// 	siteData := scraper.ScrapeOne(baseURL)

// 	if productData != nil {
// 		c.JSON(http.StatusOK, productData)
// 	} else {
// 		c.String(http.StatusNotFound, "No product data found")
// 	}
// }
