package variation

import "cloud.google.com/go/bigquery"

type Variation struct {
	ID string `bigquery:"id"`

	Identifier string `bigquery:"identifier"` // could be SKU or VariationId

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}
