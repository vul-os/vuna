package main

import (
	// "log"
	// "os"

	// // Blank-import the function package so the init() runs
	// _ "github.com/imranparuk/scraper-go"
	// "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"fmt"
	"net/http"

	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/product"
	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/product/woocommerce"
	"github.com/imranparuk/scraper-go/internal/pkg/storage"
)

func main() {
	// // Use PORT environment variable, or default to 8080.
	// port := "8080"
	// if envPort := os.Getenv("PORT"); envPort != "" {
	// 	port = envPort
	// }
	// if err := funcframework.Start(port); err != nil {
	// 	log.Fatalf("funcframework.Start: %v\n", err)
	// }
	client := http.Client{}
	var st storage.FileStorage
	productScraper := woocommerce.New(client, st)
	results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url:       "http://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/",
		Proxy: "198.211.115.186:56365",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(results)
}
