package function

import (
	"context"
	"fmt"

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
)

func init() {
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

	functions.HTTP("scraper/meta", s.Meta)
	functions.HTTP("scraper/site", s.Site)

	functions.HTTP("orchestrator/meta", o.Meta)
	functions.HTTP("orchestrator/site", o.Site)
}
