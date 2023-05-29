package scrapers

import (
	"fmt"
	"net/http"
	"net/url"

	"encoding/json"

	meta "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	product "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	shopify "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/shopify"
	woocommerce "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"

	site "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"

	utils "github.com/exolutiontech/scraper-go/internal/pkg/utils"

	// product "scraper-go/internal/pkg/scrapers/product"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/gorilla/mux"
)

type ScraperAPI struct {
	FileStorage storage.FileStorage
}

func New(
	fs storage.FileStorage,
) *ScraperAPI {
	return &ScraperAPI{
		FileStorage: fs,
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
		Url     string `json:"url"`
		Scraper string `json:"scraper"`
	}

	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}
	// Extract the proxies from the request body or query parameters
	var d jsonData
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		fmt.Println("error in json")
		http.Error(w, "Error decoding proxies", http.StatusBadRequest)
		return
	}
	productURL, err := url.QueryUnescape(d.Url)
	if err != nil {
		fmt.Println("error in producturl")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := http.Client{}
	var productScraper product.ProductScraper

	switch d.Scraper {
	case "woocommerce":
		productScraper = woocommerce.New(proxyConfig, client, api.FileStorage)
	case "shopify":
		productScraper = shopify.New(proxyConfig, client, api.FileStorage)
	default:
		fmt.Println("Scraper not implimented")
		http.Error(w, "scraper type not implimented", http.StatusBadRequest)
		return
	}
	scrapeOneResponse, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url: productURL,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("scrape one error: %s", err), http.StatusBadRequest)
		return
	}
	// Convert the product data to JSON
	resultJson, err := json.Marshal(scrapeOneResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(resultJson)
}
