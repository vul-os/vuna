package main

import (
	"fmt"
	"net/http"
	metaScraper "scraper-go/internal/pkg/scrapers/meta"
	siteScraper "scraper-go/internal/pkg/scrapers/site"

	"scraper-go/internal/pkg/scrapers/product"
	woocommerceScraper "scraper-go/internal/pkg/scrapers/product/woocommerce"
	"time"
)

func main() {
	start := time.Now()
	client := http.Client{}

	url := "https://www.biltongandbudz.co.za/product/barneys-farm-runtz-fem-autoflower/"
	proxylist := []string{""}
	scraper := woocommerceScraper.New(proxylist, client)
	data, err := scraper.ScrapeOne(product.ScrapeOneRequest{Url: url})
	if err != nil {
		fmt.Println("Error scraping product:", err)
		return
	}
	for _, data := range data.Results {
		fmt.Println(data)
	}
	elapsed := time.Since(start)
	fmt.Println("Elapsed time product:", elapsed)

	start = time.Now()
	sscraper := siteScraper.New(&client) // woocommerceScraper.New(proxylist, client), siteScraper.New()
	datas := sscraper.ScrapeOne("https://pharaohscrypt.co.za/")
	fmt.Println(datas)
	elapsed = time.Since(start)
	fmt.Println("Elapsed time site:", elapsed)

	start = time.Now()
	mscraper := metaScraper.New(&client) // woocommerceScraper.New(proxylist, client), siteScraper.New()
	_ = mscraper.ScrapeOne("https://pharaohscrypt.co.za/")
	// fmt.Println(datam)
	elapsed = time.Since(start)
	fmt.Println("Elapsed time meta:", elapsed)
	// 	"https://pharaohscrypt.co.za/")
	// fmt.Println(len(urls))
	// elapsed := time.Since(start)
	// fmt.Println("Elapsed time:", elapsed)
	// scraper =
	// info := metaScraper.MetaScrapeOne("")
	// fmt.Println(info)
	// elapsed := time.Since(start)
	// fmt.Println("Elapsed time:", elapsed)
}
