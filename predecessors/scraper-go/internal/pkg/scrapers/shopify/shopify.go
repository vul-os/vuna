package shopify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"scraper-go/internal/pkg/product"
)

type ScrapeOneRequest struct {
	Product product.Product `json:"product"`
	Proxys  []string        `json:"proxys"`
}

func ScrapeOne(w http.ResponseWriter, r *http.Request) {
	var rq ScrapeOneRequest

	err := json.NewDecoder(r.Body).Decode(&rq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	baseUrl := strings.TrimSuffix(rq.Product.Url, "/")

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Select a random SOCKS5 proxy
	proxyURL, err := url.Parse(rq.Proxys[rand.Intn(len(rq.Proxys))])
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return
	}

	// Create a HTTP transport that uses the selected SOCKS5 proxy
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Create a HTTP client that uses the custom transport
	client := &http.Client{
		Transport: tr,
	}

	// Make an HTTP GET request to retrieve the JSON data
	resp, err := client.Get(baseUrl)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Parse the JSON data into a ProductJSON struct
	var productJSON ProductJSON
	err = json.Unmarshal(body, &productJSON)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	// Print some information about the product
	product := productJSON.Product
	fmt.Println("Product ID:", product.ID)
	fmt.Println("Title:", product.Title)
	fmt.Println("Vendor:", product.Vendor)
	fmt.Println("Product Type:", product.ProductType)

	// Print the list of variants
	fmt.Println("Variants:")
	for _, variant := range product.Variants {
		fmt.Println("- ID:", variant.ID)
		fmt.Println("  Title:", variant.Title)
		fmt.Println("  Price:", variant.Price)
		fmt.Println("  SKU:", variant.SKU)
	}

	// Print the list of options
	fmt.Println("Options:")
	for _, option := range product.Options {
		fmt.Println("- Name:", option.Name)
		fmt.Println("  Values:", option.Values)
	}

	// Print the list of images
	fmt.Println("Images:")
	for _, image := range product.Images {
		fmt.Println("- ID:", image.ID)
		fmt.Println("  Src:", image.Src)
	}
}
