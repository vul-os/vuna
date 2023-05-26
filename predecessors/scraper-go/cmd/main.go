package main

import (
	// "log"
	// "os"

	// // Blank-import the function package so the init() runs
	// _ "github.com/imranparuk/scraper-go"
	// "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"fmt"
	"net/http"

	// "time"

	// "github.com/imranparuk/scraper-go/internal/pkg/orchestrator/proxy"
	// "github.com/imranparuk/scraper-go/internal/pkg/utils"
	utils "github.com/imranparuk/scraper-go/internal/pkg/utils"

	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/meta"
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
	client := http.Client{}
	proxyConfig := utils.ProxyConfig{
		Address:  "p.webshare.io:80",
		Username: "qnfhspsk-rotate",
		Password: "t62qs3cx4b6c",
	}
	var st storage.FileStorage
	productScraper := woocommerce.New(proxyConfig, client, st)
	results, err := productScraper.ScrapeOne(product.ScrapeOneRequest{
		Url: "https://www.biltongandbudz.co.za/product/sour-tropicanna-regular-12-seeds/",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("results: ", results)

	meta := meta.New(&client, st)
	data, err := meta.ScrapeOne("https://toykingdom.co.za")
	fmt.Println(len(data))
	fmt.Println(err)

}
