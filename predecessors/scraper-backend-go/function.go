package main

import (
	"log"

	"context"

	"cloud.google.com/go/bigquery"
	"github.com/exolutionza/scraper-backend-go/internal/bi"
	"github.com/gin-gonic/gin"

	// "google.golang.org/api/option"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("Router", router)
}

func router(w http.ResponseWriter, r *http.Request) {
	// Create a BigQuery client
	ctx := context.Background()
	projectID := "scraping-is-hard"

	// Provide the path to the keyfile.json
	// client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile("keyfile.json"))
	client, err := bigquery.NewClient(ctx, projectID)

	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Create a BigQuery processor
	processor := bi.NewBigQueryProcessor(client)

	// Create a Gin router
	router := gin.Default()

	// Define the routes
	router.POST("/execute", processor.TemplateAndExecuteOne)

	// // Start the server
	// port := ":8080"
	// log.Printf("Server running on port %s", port)
	// err = router.Run(port)
	// if err != nil {
	// 	log.Fatalf("Failed to start server: %v", err)
	// }
}
