package scrapers

import (
	"net/http"
	"net/url"

	"encoding/json"

	meta "github.com/imranparuk/scraper-go/internal/pkg/scrapers/meta"
	product "github.com/imranparuk/scraper-go/internal/pkg/scrapers/product"
	woocommerce "github.com/imranparuk/scraper-go/internal/pkg/scrapers/product/woocommerce"
	site "github.com/imranparuk/scraper-go/internal/pkg/scrapers/site"

	utils "github.com/imranparuk/scraper-go/internal/pkg/utils"

	// product "scraper-go/internal/pkg/scrapers/product"

	"github.com/gorilla/mux"
	"github.com/imranparuk/scraper-go/internal/pkg/storage"
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

func (api *ScraperAPI) Meta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vUrl := vars["url"]
	baseURL := "https://" + vUrl

	client := http.Client{}
	scraper := meta.New(&client, api.FileStorage)
	metaData, err := scraper.ScrapeOne(baseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.DataToJson(len(metaData), w, http.StatusOK)
}

func (api *ScraperAPI) Site(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vUrl := vars["url"]
	baseURL := "https://" + vUrl

	client := http.Client{}
	scraper := site.New(&client, api.FileStorage)
	siteData, err := scraper.ScrapeOne(baseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.DataToJson(siteData, w, http.StatusOK)
}

func (api *ScraperAPI) Product(w http.ResponseWriter, r *http.Request) {
	type jsonData struct {
		Url       string   `json:"url"`
		Proxy     string   `json:"proxy"`
		Scraper   string   `json:"scraper"`
	}
	// Extract the proxies from the request body or query parameters
	var d jsonData
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		http.Error(w, "Error decoding proxies", http.StatusBadRequest)
		return
	}
	productURL, err := url.QueryUnescape(d.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := http.Client{}
	var productScraper product.ProductScraper

	switch d.Scraper {
	case "woocommerce":
		productScraper = woocommerce.New(client, api.FileStorage)
	default:
		http.Error(w, "scraper type not implimented", http.StatusBadRequest)
		return
	}
	productData, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url:       productURL,
		Proxy: d.Proxy,
	})
	if err != nil {
		http.Error(w, "scrape one error", http.StatusBadRequest)
		return
	}
	if productData != nil {
		// Convert the product data to JSON
		productDataJSON, err := json.Marshal(productData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the response content type to application/json
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON response
		w.Write(productDataJSON)
	} else {
		http.Error(w, "No product data found", http.StatusNotFound)
	}
}
