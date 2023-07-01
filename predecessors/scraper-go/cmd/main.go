package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
)

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

func main() {
	linksFilePath := "links.txt"

	config := map[string]string{
		"product_title":           "h1.product_title.entry-title",
		"add_to_cart_input":       "input.qty",
		"add_to_cart_button":      "button.single_add_to_cart_button",
		"form_variations":         "ul[data-attribute_name='attribute_pa_size']",
		"summary_div":             "div.woocommerce-product-details__short-description",
		"price_amount":            "span.woocommerce-Price-amount",
		"sku":                     ".sku_wrapper .sku",
		"max_qty":                 "input.qty",
		"quantity_input":          "input.qty",
		"data_product_variations": ".variations_form",
		"availability_html":       "p.stock",
		"display_price":           "span.woocommerce-Price-currencySymbol",
		"variation_sku":           ".variation-sku",
		"variation_id":            ".variation-id",
		"image_src":               "img.wp-post-image",
		"attributes":              ".product_meta",
	}

	rand.Seed(time.Now().UnixNano())

	client := http.Client{}

	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}

	var st storage.FileStorage
	productScraper := woocommerce.New(proxyConfig, client, st)

	file, err := os.Open(linksFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		storeURL := strings.TrimSpace(scanner.Text())

		sc := site.New(&client, st)
		a, err := sc.ScrapeOne(storeURL)
		fmt.Println("Site Scrape Result:")
		fmt.Println("********************")
		fmt.Println("Site Data:", a)
		fmt.Println("Error:", err)

		metaScraper := meta.New(&client, st)
		data, err := metaScraper.ScrapeOne(storeURL)
		fmt.Println("Meta Scrape Result:")
		fmt.Println("********************")
		fmt.Println("Data Length:", len(data))
		fmt.Println("Error:", err)

		if len(data) == 0 || err != nil {
			fmt.Println("Skipping store due to no data or error:", storeURL)
			continue
		}

		randomSample := randomSampleSlice(data, 5)
		for _, l := range randomSample {
			results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
				Url:        l,
				FullScrape: true,
				Config:     config,
			})

			fmt.Println(fmt.Sprintf("Scraping Product: %s", l))
			fmt.Println("***********************************************")
			if err != nil {
				fmt.Println("Error occurred during product scraping:", err)
				continue
			}

			for i, pd := range results.ProductData {
				fmt.Println("ProductData:", i)
				fmt.Println("-----------------------------")
				fmt.Println("Name:", pd.Name)
				fmt.Println("Description:", pd.Description)
				fmt.Println("ImageURLs:", pd.ImageURLs)
				fmt.Println("Attributes:", pd.Attributes)
				fmt.Println("Categories:", pd.Categories)
				fmt.Println("Tags:", pd.Tags)
				fmt.Println("ProductIdentifier:", pd.ProductIdentifier)
				fmt.Println("ProductID:", pd.ProductID)
				fmt.Println("VariationID:", pd.VariationID)
				fmt.Println("SKU:", pd.SKU)
				fmt.Println("SiteIdentifier:", pd.SiteIdentifier)
				fmt.Println("DateCreated:", pd.DateCreated)
				fmt.Println()
			}

			for i, dp := range results.DataPoint {
				fmt.Println("DataPoint:", i)
				fmt.Println("-----------------------------")
				fmt.Println("ProductIdentifier:", dp.ProductIdentifier)
				fmt.Println("ProductID:", dp.ProductID)
				fmt.Println("VariationID:", dp.VariationID)
				fmt.Println("SKU:", dp.SKU)
				fmt.Println("Price:", dp.Price)
				fmt.Println("MaxQty:", dp.MaxQty)
				fmt.Println("DateCreated:", dp.DateCreated)
				fmt.Println()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
