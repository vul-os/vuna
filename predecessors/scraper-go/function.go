package function

import (
	"context"
	"fmt"
	"net/http"
    "cloud.google.com/go/bigquery"

	// metaScraper "scraper-go/internal/pkg/scrapers/meta"
	// siteScraper "scraper-go/internal/pkg/scrapers/site"
	// woocommerceScraper "scraper-go/internal/pkg/scrapers/product/woocommerce"
	// "scraper-go/internal/pkg/scrapers/product"
	orchestrator "github.com/exolutiontech/scraper-go/internal/pkg/orchestrator"
	scrapers "github.com/exolutiontech/scraper-go/internal/pkg/scrapers"

	gcsStorage "github.com/exolutiontech/scraper-go/internal/pkg/storage/gcs"

	tasks "github.com/exolutiontech/scraper-go/internal/pkg/orchestrator/tasks"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/gorilla/mux"
)

func init() {
	functions.HTTP("Router", router)
}

// router sets up the mux router and handles the HTTP request.
func router(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	rtr := mux.NewRouter()
	bucketName := "exolution-scraper-data"
	projectId := "scraping-is-hard"
	location := "us-central1"
	noRepeatQueue := "orchestrator"
	repeastQueue := "scraper"
	bigInstanceTargetUrl := "https://function-go-big-gizrqdvcaq-uc.a.run.app"
	smallInstanceTargetUrl := "https://function-go-gizrqdvcaq-uc.a.run.app"

	datasetId := "scrapers"
	datapointTableName := "datapoint_raw"
	productTableName := "product_raw"

	detailsMap := map[string]tasks.TaskCreatorDetails{
		"site": {
			TargetUrl: bigInstanceTargetUrl,
			ProjectID: projectId,
			Location:  location,
			QueueID:   repeastQueue,
		},
		"meta": {
			TargetUrl: bigInstanceTargetUrl,
			ProjectID: projectId,
			Location:  location,
			QueueID:   repeastQueue,
		},
		"product": {
			TargetUrl: smallInstanceTargetUrl,
			ProjectID: projectId,
			Location:  location,
			QueueID:   repeastQueue,
		},
		"orchestrateProductMeta": {
			TargetUrl: bigInstanceTargetUrl,
			ProjectID: projectId,
			Location:  location,
			QueueID:   noRepeatQueue,
		},
		"orchestrateProduct": {
			TargetUrl: bigInstanceTargetUrl,
			ProjectID: projectId,
			Location:  location,
			QueueID:   noRepeatQueue,
		},
	}

	bigqueryClient, err := bigquery.NewClient(ctx, projectId)
    if err != nil {
		fmt.Sprintf("Failed to create client: %v", err)
		return
    }
    
    fmt.Printf("Client created for project: %s\n", projectId)

    // Remember to close the client!
    defer bigqueryClient.Close()

	client, err := storage.NewClient(context.Background())
	if err != nil {
		fmt.Println("Error creating storage client")
		return
	}
	storage := gcsStorage.New(bucketName, *client)
	taskCreator, err := tasks.New(detailsMap)
	if err != nil {
		fmt.Println("Error creating tack creator")
		return
	}

	// Get the dataset handle.
	dataset := bigqueryClient.Dataset(datasetId)
	datapointTable := dataset.Table(datapointTableName)
	productDataTable := dataset.Table(productTableName)

	s := scrapers.New(storage, datapointTable, productDataTable)
	o := orchestrator.New(*taskCreator, storage, *bigqueryClient)

	rtr.HandleFunc("/", helloWorld).Methods(http.MethodGet)

	rtr.HandleFunc("/orchestrator/meta", o.Meta).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/site", o.Site).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/product", o.AllProducts).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/product/{siteIdentifier}", o.Product).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/product-meta", o.AllMetaProducts).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/product-meta/{file}", o.MetaProduct).Methods(http.MethodGet)

	rtr.HandleFunc("/scraper/meta/{url}", s.Meta).Methods(http.MethodGet)
	rtr.HandleFunc("/scraper/site/{url}", s.Site).Methods(http.MethodGet)
	rtr.HandleFunc("/scraper/product", s.Product).Methods(http.MethodPost)

	// Pass the HTTP request to the router
	rtr.ServeHTTP(w, r)
}

// helloWorld writes "Hello, World!" to the HTTP response.
func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
