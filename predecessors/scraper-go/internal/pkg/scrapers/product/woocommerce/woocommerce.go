package woocommerce

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"

	"strings"

	"github.com/PuerkitoBio/goquery"
)

type scraper struct {
	ProxyConfig utils.ProxyConfig
	Client      http.Client
	FileStorage storage.FileStorage
}

func New(
	pc utils.ProxyConfig,
	client http.Client,
	fs storage.FileStorage,
) product.ProductScraper {
	return &scraper{
		ProxyConfig: pc,
		Client:      client,
		FileStorage: fs,
	}
}

func (s *scraper) ScrapeOne(request product.ScrapeOneRequest) (*product.ScrapeOneResponse,
	error) {
	body, err := utils.FetchWithProxy(s.ProxyConfig, request.Url)
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
	var dataPointList []product.DataPoint

	if doc.Find("form.variations_form").Length() > 0 {
		productDataList, dataPointList, err = scrapeProductWithVariations(request.Url,
			productID, productName, doc)
		if err != nil {
			return nil, err
		}
	} else {
		productData, dataPoint, err := scrapeProductWithoutVariations(request.Url,
			productID, productName, doc)
		if err != nil {
			return nil, err
		}
		productDataList = append(productDataList, productData)
		dataPointList = append(dataPointList, dataPoint)
	}

	err = product.Save(dataPointList, productDataList, s.FileStorage, request.Url, request.FullScrape)
	if err != nil {
		return nil, err
	}
	return &product.ScrapeOneResponse{DataPoint: dataPointList,
		ProductData: productDataList}, nil
}

func scrapeProductWithoutVariations(productURL, productID, productName string, 
		doc *goquery.Document) (product.ProductData, product.DataPoint, error) {

	summaryDiv := doc.Find("div.summary")
	price := summaryDiv.Find("span.woocommerce-Price-amount.amount").Text()
	sku := summaryDiv.Find("span.sku").Text()
	maxQty := summaryDiv.Find("p.stock").Text()

	priceFloat, err := utils.PriceToFloat(price)
	if err != nil {
		return product.ProductData{}, product.DataPoint{}, err
	}
	maxQtyInt, err := utils.MaxQtyToInt(maxQty)
	if err != nil {
		inputField := doc.Find("input[name=quantity]")
		maxAttr := inputField.AttrOr("max", "0")
		maxQtyInt, err = utils.MaxQtyToInt(maxAttr)
		if err != nil {
			return product.ProductData{}, product.DataPoint{}, err
		}
	}

	otherStringIds := []string{fmt.Sprintf("%v", sku), "default"}
	hostIdentifier, productIdentifier, err := utils.StringToIdentifier(productURL, otherStringIds)
	if err != nil {
		return product.ProductData{}, product.DataPoint{}, err
	}

	imageURL, _ := doc.
		Find("div.woocommerce-product-gallery__image img").Attr("src")

	createdAt := time.Now()

	productData := product.ProductData{
		Name:        productName,
		Description: "",
		ImageURLs:   []string{imageURL},
		Attributes:  []string{},

		URL:         productURL,
		SKU:         sku,
		ProductID:   productID,
		VariationID: "",

		ProductIdentifier: productIdentifier,
		SiteIdentifier:    hostIdentifier,

		DateCreated: createdAt,
	}

	dataPoint := product.DataPoint{
		ProductIdentifier: productIdentifier,

		SKU:         sku,
		ProductID:   productID,
		VariationID: "",

		Price:  priceFloat,
		MaxQty: maxQtyInt,

		DateCreated: createdAt,
	}

	return productData, dataPoint, nil
}

func scrapeProductWithVariations(productURL, productID, productName string, 
		doc *goquery.Document) ([]product.ProductData, []product.DataPoint, error) {

	productDataList := []product.ProductData{}
	dataPointList := []product.DataPoint{}

	productVariations := doc.Find("form.variations_form").
		AttrOr("data-product_variations", "")
	variationsData := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(productVariations), &variationsData); err != nil {
		fmt.Println("Error decoding variations data:", err)
		return productDataList, dataPointList, err
	}

	for _, variation := range variationsData {
		availabilityHTML := variation["availability_html"]
		displayPrice := variation["display_price"]
		sku := variation["sku"]
		variationID := variation["variation_id"]
		priceFloat, errp := utils.PriceToFloat(displayPrice)
		maxQtyInt, errq := utils.MaxQtyToInt(availabilityHTML)
		
		otherStringIds := []string{fmt.Sprintf("%v", sku), fmt.Sprintf("%v", variationID)}
		hostIdentifier, productIdentifier, err := utils.StringToIdentifier(productURL, otherStringIds)
		if err != nil {
			continue
		}

		if errp != nil || errq != nil {
			fmt.Println("Error MaxQty,Price: ", priceFloat, maxQtyInt)
			continue
		}

		imageURL := variation["image"].(map[string]interface{})["src"]
		attributes := variation["attributes"].(map[string]interface{})
		firstValue := ""

		createdAt := time.Now()

		for _, value := range attributes {
			firstValue = value.(string)
			break
		}
		productData := product.ProductData{
			Name:        firstValue,
			Description: "",

			ImageURLs:  []string{imageURL.(string)},
			Attributes: []string{},

			URL:         productURL,
			SKU:         fmt.Sprintf("%v", sku),
			ProductID:   productID,
			VariationID: fmt.Sprintf("%v", variationID),

			ProductIdentifier: productIdentifier,
			SiteIdentifier:    hostIdentifier,

			DateCreated: createdAt,
		}

		dataPoint := product.DataPoint{
			ProductIdentifier: productIdentifier,

			SKU:         fmt.Sprintf("%v", sku),
			ProductID:   productID,
			VariationID: fmt.Sprintf("%v", variationID),

			Price:  priceFloat,
			MaxQty: maxQtyInt,

			DateCreated: createdAt,
		}

		dataPointList = append(dataPointList, dataPoint)
		productDataList = append(productDataList, productData)
	}

	return productDataList, dataPointList, nil
}
