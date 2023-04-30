package variation

import "cloud.google.com/go/bigquery"

type Variation struct {
	Id  string `bigquery:"id"`
	Sku string `bigquery:"sku"`

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}
