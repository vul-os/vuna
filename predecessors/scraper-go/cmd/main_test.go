package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery" // Import the BigQuery package
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/shopify"
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

func mainTest() {
	// Set your store URL and product link here
	storeURL := "https://blckvapour.co.za"                                                                                     // Replace with your store URL
	productLink := "https://blck.co.za/collections/propylene-glycol-pg-vegetable-glycerine-vg/products/vg-vegetable-glycerine" // Replace with your product link
	projectId := "scraping-is-hard"
	ctx := context.Background()

	bigqueryClient, err := bigquery.NewClient(ctx, projectId)
	if err != nil {
		fmt.Sprintf("Failed to create client: %v", err)
		return
	}

	datasetId := "scrapers"
	datapointTableName := "datapoint_raw"
	productTableName := "product_raw"

	// Get the dataset handle.
	dataset := bigqueryClient.Dataset(datasetId)
	datapointTable := dataset.Table(datapointTableName)
	productDataTable := dataset.Table(productTableName)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}

	productScraper := shopify.New(proxyConfig, *client, datapointTable, productDataTable)

	// Site Scrape
	scrapeSite(client, proxyConfig, storeURL)

	// Product Scrape
	scrapeProduct(client, proxyConfig, productLink, productScraper, nil)

	fmt.Println("Scraping completed.")
}
