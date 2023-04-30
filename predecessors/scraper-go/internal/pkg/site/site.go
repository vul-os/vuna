package site

import "cloud.google.com/go/bigquery"


type Site struct {
    Id      string               `bigquery:"id"`
	Url         string            `bigquery:"url"`
    Name        string          `bigquery:"name"`

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}
