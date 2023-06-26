package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
)

// Function to generate a random sample or slice
func randomSampleSlice(data []string, size int) []string {
	result := make([]string, size)

	for i := 0; i < size; i++ {
		randomIndex := rand.Intn(len(data))
		result[i] = data[randomIndex]
	}

	return result
}

func main() {
	config := map[string]string{
		"product_title":           "h1.product_title",
		"add_to_cart_input":       "input[name='add-to-cart']",
		"add_to_cart_button":      "button[name='add-to-cart']",
		"form_variations":         "form.variations_form",
		"summary_div":             "div.summary",
		"price_amount":            "span.woocommerce-Price-amount.amount",
		"sku":                     "span.sku",
		"max_qty":                 "p.stock",
		"quantity_input":          "input[name=quantity]",
		"data_product_variations": ".variations_form",
		"availability_html":       "availability_html",
		"display_price":           "display_price",
		"variation_sku":           "sku",
		"variation_id":            "variation_id",
		"image_src":               "image.src",
		"attributes":              "attributes",
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

	sphaurl := "https://dragontown.co.za"

	sc := site.New(&client, st)
	a, err := sc.ScrapeOne(sphaurl)
	fmt.Println("Site Scrape Result:")
	fmt.Println("********************")
	fmt.Println("Site Data:", a)
	fmt.Println("Error:", err)

	metaScraper := meta.New(&client, st)
	data, err := metaScraper.ScrapeOne(sphaurl)
	fmt.Println("Meta Scrape Result:")
	fmt.Println("********************")
	fmt.Println("Data Length:", len(data))
	fmt.Println("Error:", err)

	randomSample := randomSampleSlice(data, 5)
	for _, l := range randomSample {
		results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
			Url:        l,
			FullScrape: true,
			Config:     config,
		})

		fmt.Println(fmt.Sprintf("Scraping Product: %s", l))
		fmt.Println("***********************************************")
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
		fmt.Println("Error:", err)
		fmt.Println()
		fmt.Println()
	}
}
