package main

import (
	// "log"
	// "os"

	// // Blank-import the function package so the init() runs
	// _ "github.com/exolutiontech/scraper-go"
	// "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"fmt"
	"net/http"
	"os"

	// "time"

	// "github.com/exolutiontech/scraper-go/internal/pkg/orchestrator/proxy"
	// "github.com/exolutiontech/scraper-go/internal/pkg/utils"
	utils "github.com/exolutiontech/scraper-go/internal/pkg/utils"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/meta"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/woocommerce"
	// "github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product/shopify"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	gcsutils "github.com/exolutiontech/scraper-go/internal/pkg/storage/gcs"
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
	client := http.Client{}
	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}
	var st storage.FileStorage
	productScraper := woocommerce.New(proxyConfig, client, st)
	results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url:        "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/",
		FullScrape: true,
	})
	if err != nil {
		fmt.Println(err)
	}
	d, err := utils.ToMap(results.ProductData)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("results: ", d)

	// Write data to a CSV file
	csvFile, err := os.Create("data.csv")
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer csvFile.Close()

	err = gcsutils.WriteCSVFile(csvFile, d)
	if err != nil {
		fmt.Println("Error writing CSV file:", err)
	}

	meta := meta.New(&client, st)
	data, err := meta.ScrapeOne("https://3dprintingstore.co.za/")
	fmt.Println(len(data))
	fmt.Println(err)
	sc := site.New(&client, st)
	a, err := sc.ScrapeOne("https://3dprintingstore.co.za/")
	fmt.Println(a)
}
