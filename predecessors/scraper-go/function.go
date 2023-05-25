package function

import (
	"context"
	"fmt"
	"net/http"

	// metaScraper "scraper-go/internal/pkg/scrapers/meta"
	// siteScraper "scraper-go/internal/pkg/scrapers/site"
	// woocommerceScraper "scraper-go/internal/pkg/scrapers/product/woocommerce"
	// "scraper-go/internal/pkg/scrapers/product"
	orchestrator "github.com/imranparuk/scraper-go/internal/pkg/orchestrator"
	scrapers "github.com/imranparuk/scraper-go/internal/pkg/scrapers"

	gcsStorage "github.com/imranparuk/scraper-go/internal/pkg/storage/gcs"

	tasks "github.com/imranparuk/scraper-go/internal/pkg/orchestrator/tasks"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/gorilla/mux"
)

func init() {
	functions.HTTP("Router", router)
}

// router sets up the mux router and handles the HTTP request.
func router(w http.ResponseWriter, r *http.Request) {
	rtr := mux.NewRouter()

	bucketName := "exolution-scraper-data"
	targetUrl := "https://function-go-gizrqdvcaq-uc.a.run.app"
	projectId := "scraping-is-hard"
	location := "us-central1"
	queueId := "scraper"

	client, err := storage.NewClient(context.Background())
	if err != nil {
		fmt.Println("Error creating storage client")
		return
	}
	storage := gcsStorage.New(bucketName, *client)
	taskCreator, err := tasks.New(projectId, location, queueId, targetUrl)
	if err != nil {
		fmt.Println("Error creating tack creator")
		return
	}
	s := scrapers.New(storage, []string{})
	o := orchestrator.New(*taskCreator, storage, targetUrl)

	rtr.HandleFunc("/", helloWorld).Methods(http.MethodGet)

	rtr.HandleFunc("/orchestrator/meta", o.Meta).Methods(http.MethodGet)
	rtr.HandleFunc("/orchestrator/site", o.Site).Methods(http.MethodGet)

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
