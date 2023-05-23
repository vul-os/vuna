package main

import (
	"fmt"
	// "net/http"
	metaScraper "scraper-go/internal/pkg/scrapers/meta"
	// "scraper-go/internal/pkg/scrapers/product"
	// woocommerceScraper "scraper-go/internal/pkg/scrapers/product/woocommerce"
	"time"
)

func main() {
	start := time.Now()

	// client := http.Client{}
	// url := "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/"
	// proxylist := []string{""}
	// scraper := woocommerceScraper.New(proxylist, client)
	// data, err := scraper.ScrapeOne(product.ScrapeOneRequest{Url: url})
	// if err != nil {
	// 	fmt.Println("Error scraping product:", err)
	// 	return
	// }
	// for _, data := range data.Results {
	// 	fmt.Println(data)
	// }

	urls := metaScraper.MetaScrapeOne("https://bridgebooks.co.za")
	fmt.Println(len(urls))
	elapsed := time.Since(start)
	fmt.Println("Elapsed time:", elapsed)
}
