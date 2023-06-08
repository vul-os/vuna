package router

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/bigquery"
	"github.com/exolutionza/scraper-backend-go/internal/bi"
	"github.com/gorilla/mux"

	"github.com/exolutionza/scraper-backend-go/internal/search"
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

	// Create a SearchHandler
	searchHandler := search.NewSearchHandler(client)

	// Create a Gorilla Mux router
	router := mux.NewRouter()

	// Define the routes
	router.HandleFunc("/execute", processor.TemplateAndExecuteOne).Methods("POST")
	router.HandleFunc("/", helloWorld).Methods("GET")
	router.HandleFunc("/search", searchHandler.Handler).Methods("POST")

	// Apply the enableCORS middleware to all routes
	handler := enableCORS(router)

	// Serve the HTTP requests
	handler.ServeHTTP(w, r)
}

// helloWorld writes "Hello, World!" to the HTTP response.
func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Allow preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
