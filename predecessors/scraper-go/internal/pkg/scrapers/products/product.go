package products

import "cloud.google.com/go/bigquery"

type Product struct {
	ID          string            `json:"id" bigquery:"id"`
	Url         string            `json:"url" bigquery:"name"`
	StoreId     string            `json:"store_id" bigquery:"store_id"`
	
	DateAdded   bigquery.NullDate `json:"date_added" bigquery:"date_added"`
	DateUpdated bigquery.NullDate `json:"date_updated" bigquery:"date_updated"`
}
