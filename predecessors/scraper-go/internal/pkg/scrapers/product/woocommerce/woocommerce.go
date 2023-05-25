package woocommerce

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/product"
	"github.com/imranparuk/scraper-go/internal/pkg/storage"
	"github.com/imranparuk/scraper-go/internal/pkg/utils"

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

func (s *scraper) ScrapeOne(request product.ScrapeOneRequest) (*product.ScrapeOneResponse, error) {
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

	if doc.Find("form.variations_form").Length() > 0 {
		productDataList, err = scrapeProductWithVariations(request.Url, productID, productName, doc)
		if err != nil {
			return nil, err
		}
	} else {
		productData, err := scrapeProductWithoutVariations(request.Url, productID, productName, doc)
		if err != nil {
			return nil, err
		}
		productDataList = append(productDataList, productData)
	}

	if s.FileStorage != nil {
		siteUrl, err := utils.GetBaseURL(request.Url)
		if err != nil {
			return nil, err
		}
		siteUrl = utils.RemoveURLPrefix(siteUrl)
		encodedSite := utils.EncodeURL(siteUrl)

		currentDatetime := time.Now()
		formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

		fileName := fmt.Sprintf("product/%s_%s_product.csv", encodedSite, formattedDatetime)
		err = s.FileStorage.WriteData(productDataList, fileName)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	return &product.ScrapeOneResponse{Results: productDataList}, nil
}

func scrapeProductWithoutVariations(productURL, productID, productName string, doc *goquery.Document) (product.ProductData, error) {
	summaryDiv := doc.Find("div.summary")
	price := summaryDiv.Find("span.woocommerce-Price-amount.amount").Text()
	sku := summaryDiv.Find("span.sku").Text()
	maxQty := summaryDiv.Find("p.stock").Text()
	fmt.Println("here3", price, maxQty)

	priceFloat, err := PriceToFloat(price)
	if err != nil {
		return product.ProductData{}, err
	}
	maxQtyInt, err := MaxQtyToInt(maxQty)
	if err != nil {
		return product.ProductData{}, err
	}

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

	return productData, nil
}

func scrapeProductWithVariations(productURL, productID, productName string, doc *goquery.Document) ([]product.ProductData, error) {
	productDataList := []product.ProductData{}
	productVariations := doc.Find("form.variations_form").AttrOr("data-product_variations", "")
	variationsData := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(productVariations), &variationsData); err != nil {
		fmt.Println("Error decoding variations data:", err)
		return productDataList, err
	}

	for _, variation := range variationsData {
		availabilityHTML := variation["availability_html"]
		displayPrice := variation["display_price"]
		imageURL := variation["image"].(map[string]interface{})["src"]
		sku := variation["sku"]
		variationID := variation["variation_id"]
		priceFloat, errp := PriceToFloat(displayPrice)
		maxQtyInt, errq := MaxQtyToInt(availabilityHTML)

		if errp != nil || errq != nil {
			continue
		}

		attributes := variation["attributes"].(map[string]interface{})
		firstValue := ""
		for _, value := range attributes {
			firstValue = value.(string)
			break
		}

		productData := product.ProductData{
			Name:        firstValue,
			URL:         productURL,
			ImageURLs:   []string{imageURL.(string)},
			SKU:         sku.(string),
			ProductID:   productID,
			VariationID: fmt.Sprintf("%.0f", variationID),
			Price:       priceFloat,
			MaxQty:      maxQtyInt,
		}

		productDataList = append(productDataList, productData)
	}

	return productDataList, nil
}
