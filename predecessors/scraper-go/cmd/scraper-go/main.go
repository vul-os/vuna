package main

import (
	"net/http"

	"github.com/rs/zerolog/log"

	// "context"
	"flag"
	"fmt"

	// "scraper-go/internal/pkg/datapoint"
	// dpScraper "scraper-go/internal/pkg/datapoint/scraper"
	// dpStoreApi "scraper-go/internal/pkg/datapoint/store/api"
	// dpStore "scraper-go/internal/pkg/datapoint/store/bigquery"
	"scraper-go/internal/pkg/product"
	productStoreApi "scraper-go/internal/pkg/product/store/api"
	productStore "scraper-go/internal/pkg/product/store/gorm"
	"scraper-go/internal/pkg/site"
	siteStoreApi "scraper-go/internal/pkg/site/store/api"
	siteStore "scraper-go/internal/pkg/site/store/gorm"
	"scraper-go/internal/pkg/variation"
	varitationStoreApi "scraper-go/internal/pkg/variation/store/api"
	varitationStore "scraper-go/internal/pkg/variation/store/gorm"

	// "cloud.google.com/go/bigquery"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var configFileName = flag.String("config-file-name", "config", "specify config file")

func main() {

	// get config
	config, err := GetConfig(configFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("getting config from file")
	}

	log.Info().Msg("starting server...")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// projectID := "my-project-id"
	// // datasetID := "mydataset"
	// // tableID := "mytable"
	// ctx := context.Background()
	// client, err := bigquery.NewClient(ctx, projectID)
	// if err != nil {
	// 	fmt.Errorf("bigquery.NewClient: %w", err)
	// }
	// defer client.Close()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", config.DatabaseHost,
		config.DatabaseUser, config.DatabasePassword, config.DatabaseName, config.DatabasePort)
	gormDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Gorm DB error")
	}

	err = gormDb.AutoMigrate(
		site.Site{},
		product.Product{},
		variation.Variation{},
		// datapoint.DataPoint{},
	)

	SiteStore := siteStore.New(
		gormDb,
	)

	ProductStore := productStore.New(
		gormDb,
	)

	VaritationStore := varitationStore.New(
		gormDb,
	)

	// DataPointStore := dpStore.New(
	// 	client,
	// )

	// APIs
	SiteStoreApi := siteStoreApi.New(
		SiteStore,
	)

	ProductStoreApi := productStoreApi.New(
		ProductStore,
	)

	VariationStoreApi := varitationStoreApi.New(
		VaritationStore,
	)

	// DataPointStoreApi := dpStoreApi.New(
	// 	DataPointStore,
	// )

	// DataPointScraper := dpScraper.New(
	// 	ProductStore,
	// 	VaritationStore,
	// 	DataPointStore,
	// )

	// Mount the sub-routers
	r.Route("/site", func(r chi.Router) {
		r.Mount("/", SiteStoreApi.Routes())
		// r.Route("/scrape", func(r chi.Router) {
		// 	r.Mount()
		// })
	})

	r.Route("/product", func(r chi.Router) {
		r.Mount("/", ProductStoreApi.Routes())
	})

	r.Route("/variation", func(r chi.Router) {
		r.Mount("/", VariationStoreApi.Routes())
	})

	// r.Route("/datapoint", func(r chi.Router) {
	// 	r.Mount("/", DataPointStoreApi.Routes())
	// 	r.Mount("/scraper", DataPointScraper.Routes())
	// })
	http.ListenAndServe(fmt.Sprintf("%s:%s", "localhost", config.ServerPort), r)
}
