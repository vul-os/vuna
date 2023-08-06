package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
)

type StoreResult struct {
	URL         string
	TotalPrice  float64
	TotalMaxQty int
}

func randomSampleSlice(data []string, size int) []string {
	if len(data) <= size {
		return data
	}

	result := make([]string, size)

	for i := 0; i < size; i++ {
		randomIndex := rand.Intn(len(data))
		result[i] = data[randomIndex]
		data = append(data[:randomIndex], data[randomIndex+1:]...)
	}

	return result
}

func scrapeStore(client *http.Client, st storage.FileStorage, proxyConfig utils.ProxyConfig, storeURL string, wg *sync.WaitGroup, resultsChan chan<- StoreResult, productScraper product.ProductScraper, outputWriter *bufio.Writer, config map[string]string) {
	defer wg.Done()

	sc := site.New(client, st)
	a, err := sc.ScrapeOne(storeURL)
	fmt.Println("Site Scrape Result:")
	fmt.Println("********************")
	fmt.Println("Site Data:", a)
	fmt.Println("Error:", err)

	metaScraper := meta.New(client, st)
	data, err := metaScraper.ScrapeOne(storeURL)
	fmt.Println("Meta Scrape Result:")
	fmt.Println("********************")
	fmt.Println("Data Length:", len(data))
	fmt.Println("Error:", err)

	if len(data) == 0 || err != nil {
		fmt.Println("Skipping store due to no data or error:", storeURL)
		return
	}

	randomSample := randomSampleSlice(data, 50)
	totalMaxQty := 0
	totalPrice := 0.0

	for _, l := range randomSample {
		results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
			Url:    l,
			Config: config, // Pass the config here
			Save:   false,
		})
		fmt.Println()
		fmt.Println(fmt.Sprintf("Scraping Product: %s", l))
		fmt.Println("***********************************************")
		if err != nil {
			fmt.Println("Error occurred during product scraping:", err)
			continue
		}

		fmt.Println(results)
		fmt.Println(storeURL, totalPrice, totalMaxQty)
		outputWriter.WriteString(fmt.Sprintf("%s\n", results))
		outputWriter.Flush()

		for _, dp := range results.DataPoint {
			totalMaxQty += dp.MaxQty
			totalPrice += dp.Price
		}
	}

}

func main2() {
	linksFilePath := "/home/imran/Documents/data/scraperlinks/test_iceid.txt"
	outputFilePath := "/home/imran/Documents/data/scraperlinks/wooice.txt"
	startingLine := 0

	rand.Seed(time.Now().UnixNano())

	client := &http.Client{}

	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}

	var st storage.FileStorage

	file, err := os.Open(linksFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	outputFile, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening output file:", err)
		return
	}
	defer outputFile.Close()
	outputWriter := bufio.NewWriter(outputFile)

	lineScanner := bufio.NewScanner(file)

	var wg sync.WaitGroup
	resultsChan := make(chan StoreResult)
	doneChan := make(chan struct{})

	productScraper := woocommerce.New(proxyConfig, *client, nil, nil)

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

	maxGoroutines := 32
	concurrencyLimiter := make(chan struct{}, maxGoroutines)

	for i := 0; i < startingLine; i++ {
		if !lineScanner.Scan() {
			fmt.Println("Starting line exceeds the number of lines in the file")
			return
		}
	}

	go func() {
		defer close(resultsChan)
		for lineScanner.Scan() {
			storeURL := "https://" + strings.TrimSpace(lineScanner.Text())

			concurrencyLimiter <- struct{}{}

			wg.Add(1)
			go func(storeURL string) {
				defer func() {
					<-concurrencyLimiter
				}()
				scrapeStore(client, st, proxyConfig, storeURL, &wg, resultsChan, productScraper, outputWriter, config)
			}(storeURL)
		}
		wg.Wait()
		close(doneChan)
	}()

	totalMaxQty := 0
	totalPrice := 0.0

	go func() {
		for result := range resultsChan {
			totalPrice += result.TotalPrice
			totalMaxQty++
		}
	}()

	<-doneChan

	fmt.Println("Total Stores:", totalMaxQty)
	fmt.Println("Total Price:", totalPrice)

	if err := lineScanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
