package main

import (
	"fmt"
	"net/http"
	"time"

	// Import the BigQuery package
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
)

func scrapeSite(client *http.Client, proxyConfig utils.ProxyConfig, storeURL string) {
	sc := site.New(client, nil)
	a, err := sc.ScrapeOne(storeURL)
	fmt.Println("Site Scrape Result:")
	fmt.Println("********************")
	fmt.Println("Site Data:", a)
	fmt.Println("Error:", err)
}

func scrapeProduct(client *http.Client, proxyConfig utils.ProxyConfig, productLink string, productScraper product.ProductScraper, config map[string]string) {
	results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url:    productLink,
		Config: config,
		Save:   false,
	})

	fmt.Println()
	fmt.Println(fmt.Sprintf("Scraping Product: %s", productLink))
	fmt.Println("***********************************************")
	if err != nil {
		fmt.Println("Error occurred during product scraping:", err)
	}

	fmt.Println(results)
}

func main() {
	// Set your store URL and product link here
	storeURL := "https://klopperssport.co.za"                                                         // Replace with your store URL
	productLink := "https://klopperssport.co.za/product/puma-ultra-pro-protect-rc-goalkeeper-gloves/" // Replace with your product link

	config := map[string]string{
		"product_title":           "h1.product_title",
		"add_to_cart_input":       "input[name='add-to-cart']",
		"add_to_cart_button":      "button[name='add-to-cart']",
		"form_variations":         "[data-product_variations]",
		"summary_div":             "div.summary",
		"price_amount":            "span.woocommerce-Price-amount.amount",
		"sku":                     "span.sku",
		"max_qty":                 "p.stock",
		"quantity_input":          "input[name=quantity]",
		"data_product_variations": "data-product_variations",
		"availability_html":       "availability_html",
		"display_price":           "display_price",
		"variation_sku":           "sku",
		"variation_id":            "variation_id",
		"image_src":               "image.src",
		"attributes":              "attributes",
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}

	productScraper := woocommerce.New(proxyConfig, *client, nil, nil)

	// Site Scrape
	scrapeSite(client, proxyConfig, storeURL)

	// Product Scrape
	scrapeProduct(client, proxyConfig, productLink, productScraper, config)

	fmt.Println("Scraping completed.")
}
