package router

import (
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/bigquery"
	"github.com/exolutionza/scraper-backend-go/internal/bi"
	"github.com/gorilla/mux"

	"context"
)

func Router(w http.ResponseWriter, r *http.Request) {
	// Create a BigQuery client
	ctx := context.Background()
	projectID := "scraping-is-hard"

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Create a BigQuery processor
	processor := bi.NewBigQueryProcessor(client)

	// Create a Gorilla Mux router
	router := mux.NewRouter()

	// Define the routes
	router.HandleFunc("/execute", processor.TemplateAndExecuteOne).Methods("POST")
	router.HandleFunc("/", helloWorld).Methods("GET")

	// Handle CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Apply the router
	router.ServeHTTP(w, r)
}

// helloWorld writes "Hello, World!" to the HTTP response.
func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
