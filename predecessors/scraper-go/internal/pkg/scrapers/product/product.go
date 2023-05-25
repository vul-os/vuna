package product

import (
	"errors"
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

func ToMap(p []ProductData) ([]map[string]string, error) {
	var retData []map[string]string
	for _, pd := range p {
		data := make(map[string]string)
		fmt.Println(pd, reflect.ValueOf(pd))
		v := reflect.ValueOf(pd) // Get the value of the struct
		t := v.Type()             // Get the type of the struct

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := t.Field(i).Name

			switch field.Kind() {
			case reflect.String:
				data[fieldName] = field.String()
			case reflect.Slice:
				// Check if the slice contains only strings.
				if field.Type().Elem().Kind() == reflect.String {
					data[fieldName] = ""
				} else {
					return nil, errors.New(fmt.Sprintf("Unsupported field type (slice of non-strings): %s", fieldName))
				}
			case reflect.Float64:
				data[fieldName] = fmt.Sprintf("%.2f", field.Float())
			case reflect.Int:
				data[fieldName] = fmt.Sprintf("%d", field.Int())
			default:
				return nil, errors.New(fmt.Sprintf("Unsupported field type: %s", fieldName))
			}
		}
		retData = append(retData, data)
	}

	return retData, nil
}