package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strings"
	"scraper-go/internal/pkg/scrapers/product"
)

type scraper struct {
	Url string
	ProxyList []string
	Client http.Client
}

func New(
	url string,
	proxyList []string,
	client http.Client,
) product.ProductScraper {
	return &scraper{
		Url: url,
		ProxyList: proxyList,
		Client: client,
	}
}

func (s *scraper) ScrapeOne(request product.ScrapeOneRequest) (*product.ScrapeOneResponse, error) {
	response, err := s.Client.Get(request.Url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	productName := doc.Find("h1.product_title").Text()

	productID, _ := doc.Find("input[name='add-to-cart']").Attr("value")
	if productID == "" {
		productID, _ = doc.Find("button[name='add-to-cart']").Attr("value")
	}

	var productDataList []product.ProductData

	if doc.Find("form.variations_form").Length() > 0 {
		productDataList = scrapeProductWithVariations(request.Url, productID, productName, doc)
	} else {
		productData := scrapeProductWithoutVariations(request.Url, productID, productName, doc)
		productDataList = append(productDataList, productData)
	}

	return &product.ScrapeOneResponse{Results: productDataList}, nil
}

func scrapeProductWithoutVariations(productURL, productID, productName string, doc *goquery.Document) product.ProductData {
	summaryDiv := doc.Find("div.summary")
	price := summaryDiv.Find("span.woocommerce-Price-amount.amount").Text()
	sku := summaryDiv.Find("span.sku").Text()
	maxQty := summaryDiv.Find("p.stock").Text()

	priceFloat := priceToFloat(price)
	maxQtyInt := maxQtyToInt(maxQty)

	imageURL, _ := doc.Find("div.woocommerce-product-gallery__image img").Attr("src")

	productData := product.ProductData{
		Name:        productName,
		URL:         productURL,
		ImageURLs:   []string{imageURL},
		SKU:         sku,
		ProductID:   productID,
		VariationID: "",
		Price:       priceFloat,
		MaxQty:      maxQtyInt,
	}

	return productData
}

func scrapeProductWithVariations(productURL, productID, productName string, doc *goquery.Document) []product.ProductData {
	productDataList := []product.ProductData{}
	productVariations := doc.Find("form.variations_form").AttrOr("data-product_variations", "")
	variationsData := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(productVariations), &variationsData); err != nil {
		fmt.Println("Error decoding variations data:", err)
		return productDataList
	}

	for _, variation := range variationsData {
		availabilityHTML := variation["availability_html"].(string)
		displayPrice := variation["display_price"].(string)
		imageURL := variation["image"].(map[string]interface{})["src"].(string)
		sku := variation["sku"].(string)
		variationID := variation["variation_id"].(float64)

		priceFloat := priceToFloat(displayPrice)
		maxQtyInt := maxQtyToInt(availabilityHTML)

		attributes := variation["attributes"].(map[string]interface{})
		firstValue := ""
		for _, value := range attributes {
			firstValue = value.(string)
			break
		}

		productData := product.ProductData{
			Name:        firstValue,
			URL:         productURL,
			ImageURLs:   []string{imageURL},
			SKU:         sku,
			ProductID:   productID,
			VariationID: fmt.Sprintf("%.0f", variationID),
			Price:       priceFloat,
			MaxQty:      maxQtyInt,
		}

		productDataList = append(productDataList, productData)
	}

	return productDataList
}

func priceToFloat(price string) float64 {
	// Add code to convert price string to float
	return 0.0
}

func maxQtyToInt(maxQty string) int {
	// Add code to convert maxQty string to int
	return 0
}

func main() {
	scraper := Woo()
	productURL := "https://example.com/product"
	productData, err := scraper.ScrapeProduct(productURL)
	if err != nil {
		fmt.Println("Error scraping product:", err)
		return
	}

	for _, data := range productData {
		fmt.Println(data)
	}
}
