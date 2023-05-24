package product

import (
	"fmt"
	"reflect"
)

// ProductData is a struct representing the product data
type ProductData struct {
	Name      string
	ImageURLs []string
	// Attribute    string

	URL         string
	ProductID   string
	VariationID string
	SKU         string

	Price  float64
	MaxQty int
}

type ProductScraper interface {
	ScrapeOne(ScrapeOneRequest) (*ScrapeOneResponse, error)
}

type ScrapeOneRequest struct {
	Url string
}

type ScrapeOneResponse struct {
	Results []ProductData
}

func (p *ProductData) ToMap() map[string]string {
	data := make(map[string]string)

	v := reflect.ValueOf(p).Elem() // Get the value of the struct
	t := v.Type()                  // Get the type of the struct

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		switch field.Interface().(type) {
		case string:
			data[fieldName] = field.String()
		// case []string:
		// 	data[fieldName] = strings.Join(field.Interface().([]string), ",")
		case float64:
			data[fieldName] = fmt.Sprintf("%.2f", field.Float())
		case int:
			data[fieldName] = fmt.Sprintf("%d", field.Int())
		}
	}

	return data
}
