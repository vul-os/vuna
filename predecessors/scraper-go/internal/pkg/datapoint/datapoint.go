package datapoint

import "cloud.google.com/go/bigquery"

type DataPoint struct {
	ID    string `bigquery:"id"`
	VarID string `bigquery:"var_id"`

	MaxQty int     `bigquery:"max_qty"`
	Price  float32 `bigquery:"price"`

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}
