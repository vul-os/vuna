package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

type ProductDataForBQ struct {
	Attributes       string
	Categories       string
	Tags             string
	DateCreated      time.Time
	Description      string
	ImageURLs        string
	Name             string
	ProductID        string
	ProductIdentifier string
	SKU              string
	SiteIdentifier   string
	URL              string
	VariationID      string
}

func Save(dataPointList []DataPoint, productDataList []ProductData, dataPointTable *bigquery.Table, productTable *bigquery.Table) error {
	ctx := context.Background()

	// Save datapoints to the datapoint_raw table
	var dataPointRows []bigquery.ValueSaver
	for _, dp := range dataPointList {
		dataPointRows = append(dataPointRows, &bigquery.StructSaver{
			Struct:   dp,
			InsertID: fmt.Sprintf("%d", time.Now().UnixNano()),
		})
	}

	if err := dataPointTable.Inserter().Put(ctx, dataPointRows); err != nil {
		return err
	}

	// Save product data to the product_raw table
	var productDataRows []bigquery.ValueSaver
	for _, pd := range productDataList {
		// Serialize string lists to JSON strings.
		attributesJSON, err := json.Marshal(pd.Attributes)
		if err != nil {
			return err
		}

		categoriesJSON, err := json.Marshal(pd.Categories)
		if err != nil {
			return err
		}

		tagsJSON, err := json.Marshal(pd.Tags)
		if err != nil {
			return err
		}

		imageURLsJSON, err := json.Marshal(pd.ImageURLs)
		if err != nil {
			return err
		}

		// Construct the ProductDataForBQ object for BigQuery
		productDataBQ := ProductDataForBQ{
			Attributes:       string(attributesJSON),
			Categories:       string(categoriesJSON),
			Tags:             string(tagsJSON),
			DateCreated:      pd.DateCreated,
			Description:      pd.Description,
			ImageURLs:        string(imageURLsJSON),
			Name:             pd.Name,
			ProductID:        pd.ProductID,
			ProductIdentifier: pd.ProductIdentifier,
			SKU:              pd.SKU,
			SiteIdentifier:   pd.SiteIdentifier,
			URL:              pd.URL,
			VariationID:      pd.VariationID,
		}

		productDataRows = append(productDataRows, &bigquery.StructSaver{
			Struct:   productDataBQ,
			InsertID: fmt.Sprintf("%d", time.Now().UnixNano()),
		})
	}

	if err := productTable.Inserter().Put(ctx, productDataRows); err != nil {
		return err
	}

	return nil
}
