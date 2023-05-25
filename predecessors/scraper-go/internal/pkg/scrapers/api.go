package scrapers

import (
	"encoding/json"
	"net/http"

	meta "github.com/imranparuk/scraper-go/internal/pkg/scrapers/meta"
	site "github.com/imranparuk/scraper-go/internal/pkg/scrapers/site"

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

	// Convert the meta data to JSON
	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(metaDataJSON)
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

	// Convert the site data to JSON
	siteDataJSON, err := json.Marshal(siteData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(siteDataJSON)
}

// func (api *ScraperAPI) Product(w http.ResponseWriter, r *http.Request) {
// 	productURL, err := url.QueryUnescape(r.URL.Path)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var proxies []string
// 	// Extract the proxies from the request body or query parameters

// 	var scraperCode string
// 	// Extract the scraper code from the request body or query parameters

// 	client := http.Client{}
// 	scraper := product.New(&client, api.FileStorage)
// 	productData := scraper.ScrapeOne(productURL)

// 	if productData != nil {
// 		// Convert the product data to JSON
// 		productDataJSON, err := json.Marshal(productData)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		// Set the response content type to application/json
// 		w.Header().Set("Content-Type", "application/json")

// 		// Write the JSON response
// 		w.Write(productDataJSON)
// 	} else {
// 		http.Error(w, "No product data found", http.StatusNotFound)
// 	}
// }
