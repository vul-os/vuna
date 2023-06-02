package bi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"cloud.google.com/go/bigquery"
)

type BigQueryProcessor struct {
	Client *bigquery.Client
}

func NewBigQueryProcessor(client *bigquery.Client) *BigQueryProcessor {
	return &BigQueryProcessor{
		Client: client,
	}
}

func (bp *BigQueryProcessor) TemplateAndExecuteOne(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Extract the name and template_dict from the data
	name, _ := data["name"].(string)
	templateDict, _ := data["template_dict"].(map[string]interface{})

	// Process the file contents
	fileContents := ProcessFile(name)
	if fileContents == "" {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// Apply the template substitution
	tmpl, err := template.New("query").Parse(fileContents)
	if err != nil {
		http.Error(w, "Failed to parse query template", http.StatusInternalServerError)
		return
	}

	var queryBuilder strings.Builder
	err = tmpl.Execute(&queryBuilder, templateDict)
	if err != nil {
		http.Error(w, "Failed to execute query template", http.StatusInternalServerError)
		return
	}

	query := queryBuilder.String()

	// Execute the BigQuery query
	rows, err := bp.Client.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to execute BigQuery query", http.StatusInternalServerError)
		return
	}
	defer rows.Stop()

	// Process the query results
	var results []map[string]bigquery.Value
	for {
		var result map[string]bigquery.Value
		err := rows.Next(&result)
		if err == bigquery.Done {
			break
		}
		if err != nil {
			http.Error(w, "Failed to retrieve query result", http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	// Convert the results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Failed to convert results to JSON", http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
