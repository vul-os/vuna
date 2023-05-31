package shopify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/product"
	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
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

	_, encodedSite, err := utils.UrlToIdetifier(request.Url)
	if err != nil {
		return nil, err
	}

	body, err := utils.FetchWithProxy(s.ProxyConfig, request.Url+".json")
	if err != nil {
		return nil, err
	}

	var response ProductResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var productDataList []product.ProductData = nil
	var dataPointList []product.DataPoint = nil

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

		productUrl := utils.RemoveURLPrefix(request.Url)
		encodedProductUrl, err := utils.EncodeAndCompressURL(productUrl)
		if err != nil {
			continue
		}
		productIdentifier := fmt.Sprintf("%s-%s$%d", encodedProductUrl, variant.SKU, variant.ID)

		dataPoint := product.DataPoint{
			ProductIdentifier: productIdentifier,

			SKU:         variant.SKU,
			ProductID:   fmt.Sprintf("%d", response.Product.ID),
			VariationID: fmt.Sprintf("%d", variant.ID),

			Price:  priceFloat,
			MaxQty: maxQtyInt,
		}

		dataPointList = append(dataPointList, dataPoint)

		if request.FullScrape {
			imageSrcs := make([]string, len(response.Product.Images))
			for i, img := range response.Product.Images {
				imageSrcs[i] = img.Src
			}

			productData := product.ProductData{
				Name:        response.Product.Title,
				Description: "",

				ImageURLs:  imageSrcs,
				Attributes: []string{},

				URL:         request.Url,
				SKU:         variant.SKU,
				ProductID:   fmt.Sprintf("%d", response.Product.ID),
				VariationID: fmt.Sprintf("%d", variant.ID),

				ProductIdentifier: productIdentifier,
				SiteIdentifier: encodedSite,

			}

			productDataList = append(productDataList, productData)
		}
	}

	err = product.Save(dataPointList, productDataList, s.FileStorage, request.Url, request.FullScrape)
	if err != nil {
		return nil, err
	}
	return &product.ScrapeOneResponse{ProductData: productDataList, DataPoint: dataPointList}, nil
}
