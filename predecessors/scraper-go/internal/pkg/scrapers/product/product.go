package product

import (
	"time"
)

// // ProductData is a struct representing the product data
// type ProductData struct {
// 	Name string
// 	// ImageURLs []string
// 	// Attribute    string

// 	URL         string
// 	ProductID   string
// 	VariationID string
// 	SKU         string

// 	Price  float64
// 	MaxQty int
// }

// ProductData is a struct representing the product data
type ProductData struct {
	Name        string
	Description string

	ImageURLs  []string
	Attributes []string
	Categories []string
	Tags       []string

	URL         string
	ProductID   string
	VariationID string
	SKU         string

	ProductIdentifier string
	SiteIdentifier string
}

type DataPoint struct {
	ProductIdentifier string

	ProductID   string
	VariationID string
	SKU         string

	Price  float64
	MaxQty int

	DateCreated time.Time
}

type ProductScraper interface {
	ScrapeOne(ScrapeOneRequest) (*ScrapeOneResponse, error)
}

type ScrapeOneRequest struct {
	Url        string
	FullScrape bool
}

type ScrapeOneResponse struct {
	DataPoint   []DataPoint
	ProductData []ProductData
}
