package main

import (
	// "log"
	// "os"

	// // Blank-import the function package so the init() runs
	// _ "github.com/exolutiontech/scraper-go"
	// "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"fmt"
	"net/http"

	// "os"

	// "time"

	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
	// utils "github.com/exolutiontech/scraper-go/internal/pkg/utils"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"

	// "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/shopify"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"

	// "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/shopify"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	// gcsutils "github.com/exolutiontech/scraper-go/internal/pkg/storage/gcs"
	"math/rand"
	"time"
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
	// // Use PORT environment variable, or default to 8080.
	// port := "8080"
	// if envPort := os.Getenv("PORT"); envPort != "" {
	// 	port = envPort
	// }
	// if err := funcframework.Start(port); err != nil {
	// 	log.Fatalf("funcframework.Start: %v\n", err)
	// }
	// proxyListRaw, err := proxy.CreateProxyList()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// proxyList := utils.TestProxies("biltongandbudz.co.za", proxyListRaw, time.Second*5)
	// if len(proxyList) == 0 {
	// 	fmt.Println("no list")
	// }
	// fmt.Println(proxyList)
	// return
	// return
	rand.Seed(time.Now().UnixNano())

	client := http.Client{}
	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}
	var st storage.FileStorage
	productScraper := woocommerce.New(proxyConfig, client, st)
	// results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
	// 	Url:        "https://mini-me.co.za/product/4-way-tyre-iron-silver/",
	// 	FullScrape: true,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(results)
	// d, err := utils.ToMap(results.ProductData)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("results: ", d)

	// // Write data to a CSV file
	// csvFile, err := os.Create("data.csv")
	// if err != nil {
	// 	fmt.Println("Error creating CSV file:", err)
	// 	return
	// }
	// defer csvFile.Close()

	// err = gcsutils.WriteCSVFile(csvFile, d)
	// if err != nil {
	// 	fmt.Println("Error writing CSV file:", err)
	// }
	sphaurl := "https://biltongandbudz.co.za"

	sc := site.New(&client, st)
	a, err := sc.ScrapeOne(sphaurl)
	fmt.Println(a)
	fmt.Println(err)

	meta := meta.New(&client, st)
	data, err := meta.ScrapeOne(sphaurl)
	fmt.Println(len(data))
	fmt.Println(err)
	fmt.Println(data)
	randomSample := randomSampleSlice(data, 5)
	for _, l := range randomSample {
		results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
			Url:        l,
			FullScrape: true,
			Config:     config,
		})
		fmt.Println(fmt.Sprintf("start-%s----------------------------------------------------------", l))

		for i, pd := range results.ProductData {
			fmt.Println("***********************************************")
			fmt.Println("ProductData: ", i)
			fmt.Println("Name: ", pd.Name)
			fmt.Println("Description : ", pd.Description)
			fmt.Println("ImageURLs : ", pd.ImageURLs)
			fmt.Println("Attributes : ", pd.Attributes)
			fmt.Println("Categories : ", pd.Categories)
			fmt.Println("Tags  : ", pd.Tags)
			fmt.Println("ProductIdentifier: ", pd.ProductIdentifier)
			fmt.Println("ProductId: ", pd.ProductID)
			fmt.Println("VariationID : ", pd.VariationID)
			fmt.Println("SKU  : ", pd.SKU)
			fmt.Println("SiteIdentifier  : ", pd.SiteIdentifier)
			fmt.Println("DateCreated  : ", pd.DateCreated)
			fmt.Println("***********************************************")

		}
		for i, dp := range results.DataPoint {
			fmt.Println("***********************************************")
			fmt.Println("ProductData: ", i)
			fmt.Println("ProductIdentifier: ", dp.ProductIdentifier)
			fmt.Println("ProductId: ", dp.ProductID)
			fmt.Println("VariationID : ", dp.VariationID)
			fmt.Println("SKU  : ", dp.SKU)
			fmt.Println("Price  : ", dp.Price)
			fmt.Println("MaxQty : ", dp.MaxQty)
			fmt.Println("DateCreated  : ", dp.DateCreated)
			fmt.Println("***********************************************")

		}
		fmt.Println("error: ", err)

		fmt.Println(fmt.Sprintf("start-%s----------------------------------------------------------", l))

		fmt.Println()
		fmt.Println()

		// fmt.Println(results)
	}
}
