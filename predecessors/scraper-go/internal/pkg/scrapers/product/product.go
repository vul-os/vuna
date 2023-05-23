package product

// ProductData is a struct representing the product data
type ProductData struct {
	Name         string
	ImageURLs    []string
	// Attribute    string

	URL          string
	ProductID    string
	VariationID  string
	SKU          string
	
	Price        float64
	MaxQty       int
}

type ProductScraper interface {
	ScrapeOne(ScrapeOneRequest) (*ScrapeOneResponse, error)
}

type ScrapeOneRequest struct {
	Url string
	ProxyList []string
}

type ScrapeOneResponse struct {
	Results []ProductData
}