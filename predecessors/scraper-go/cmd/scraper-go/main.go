package main

import (
	"github.com/rs/zerolog/log"

	"context"
	"fmt"
	
	productsStore "scraper-go/internal/pkg/scrapers/products/store"

	"cloud.google.com/go/bigquery"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func main() {
	log.Info().Msg("starting server...")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// // RESTy routes
	// r.Route("/store", func(r chi.Router) {
	// 	// r.Get("/delete", SearchArticles)
	// 	// r.Get("/scrape", SearchArticles)

	// 	r.Route("/{storeID}", func(r chi.Router) {
	// 		r.Get("/scrape", GetArticle)       // GET /articles/123
	// 	})
	// })

	// r.Route("/product/{productID}", func(r chi.Router) {
	// 	r.Get("/scrape", SearchArticles)
	// })

	projectID := "my-project-id"
	// datasetID := "mydataset"
	// tableID := "mytable"
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		fmt.Errorf("bigquery.NewClient: %w", err)
	}

	ProductsStore := productsStore.New(
		client,
	)

	defer client.Close()
}
