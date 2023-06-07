package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type SearchQuery struct {
	Query string `json:"query"`
}

type SearchResult map[string]bigquery.Value

type SearchHandler struct {
	client *bigquery.Client
}

func NewSearchHandler(client *bigquery.Client) *SearchHandler {
	return &SearchHandler{
		client: client,
	}
}

const PRODUCT_SQL = "SELECT * FROM `scraping-is-hard.scrapers.product_unique` WHERE SEARCH((name, url, sku), '%s') LIMIT 15"
const SITE_SQL = "SELECT * FROM `scraping-is-hard.scrapers.site_unique` WHERE SEARCH((name, url), '%s') LIMIT 10"

func (sh *SearchHandler) PerformSearch(query string, sql string) ([]bigquery.Value, error) {
	ctx := context.Background()

	// Create the BigQuery query for the product_index table
	q := fmt.Sprintf(sql, query)
	fmt.Println(q)

	// Run the BigQuery query for the product_index table
	itProduct, err := sh.client.Query(q).Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []bigquery.Value

	// Iterate over the query results and collect them
	for {
		row := make(map[string]bigquery.Value)
		err := itProduct.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		results = append(results, row)
	}

	return results, nil
}

func (sh *SearchHandler) Handler(w http.ResponseWriter, r *http.Request) {
	// Parse the search query from the request body
	var searchQuery SearchQuery
	err := json.NewDecoder(r.Body).Decode(&searchQuery)
	if err != nil {
		http.Error(w, "Failed to parse search query", http.StatusBadRequest)
		return
	}

	// Perform the search using the SearchHandler for the product_index table
	resultsProduct, err := sh.PerformSearch(searchQuery.Query, PRODUCT_SQL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to perform product search", http.StatusInternalServerError)
		return
	}

	// Perform the search using the SearchHandler for the site_index table
	resultsSite, err := sh.PerformSearch(searchQuery.Query, SITE_SQL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to perform site search", http.StatusInternalServerError)
		return
	}

	// Merge the search results from both tables
	results := append(resultsProduct, resultsSite...)

	// Write the search results to the response
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		http.Error(w, "Failed to write search results", http.StatusInternalServerError)
		return
	}
}
