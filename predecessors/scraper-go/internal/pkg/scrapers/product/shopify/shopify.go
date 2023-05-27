package shopify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/product"
	"github.com/imranparuk/scraper-go/internal/pkg/storage"
	"github.com/imranparuk/scraper-go/internal/pkg/utils"
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

type ProductResponse struct {
	Product struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"body_html"`
		Vendor      string `json:"vendor"`
		Price       string `json:"price"`
		Images      []struct {
			Src string `json:"src"`
		} `json:"images"`
		Variants []struct {
			ID                int64  `json:"id"`
			SKU               string `json:"sku"`
			Price             string `json:"price"`
			InventoryQuantity int    `json:"inventory_quantity"`
		} `json:"variants"`
	} `json:"product"`
}

func (s *scraper) ScrapeOne(request product.ScrapeOneRequest) (*product.ScrapeOneResponse, error) {
	body, err := utils.FetchWithProxy(s.ProxyConfig, request.Url+".json")
	if err != nil {
		return nil, err
	}

	var response ProductResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var productDataList []product.ProductData

	for _, variant := range response.Product.Variants {
		priceFloat, err := utils.PriceToFloat(variant.Price)
		if err != nil {
			// Handle error for price conversion
			log.Println("Error converting price:", err)
			continue
		}

		maxQtyInt, err := utils.MaxQtyToInt(variant.InventoryQuantity)
		if err != nil {
			// Handle error for max quantity conversion
			log.Println("Error converting max quantity:", err)
			continue
		}

		productData := product.ProductData{
			Name:        utils.CleanString(response.Product.Title),
			URL:         utils.CleanString(request.Url),
			SKU:         utils.CleanString(variant.SKU),
			ProductID:   fmt.Sprintf("%d", response.Product.ID),
			VariationID: fmt.Sprintf("%d", variant.ID),
			Price:       priceFloat,
			MaxQty:      maxQtyInt,
		}
		productDataList = append(productDataList, productData)
	}

	siteURL, err := utils.GetBaseURL(request.Url)
	if err != nil {
		return nil, err
	}

	encodedSite := utils.EncodeURL(siteURL)

	currentDatetime := time.Now()
	formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

	fileName := fmt.Sprintf("product/%s_%s_product.csv", encodedSite, formattedDatetime)

	pdl, err := product.ToMap(productDataList)
	if err != nil {
		return &product.ScrapeOneResponse{Results: nil}, err
	}

	if s.FileStorage != nil && len(pdl) > 0 {
		err = s.FileStorage.WriteData(pdl, fileName)
		if err != nil {
			return &product.ScrapeOneResponse{Results: nil}, err
		}
	}

	return &product.ScrapeOneResponse{Results: productDataList}, nil
}
