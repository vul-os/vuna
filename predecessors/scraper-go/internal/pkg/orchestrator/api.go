package orchestrator

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/bigquery"

	"github.com/gorilla/mux"

	"github.com/exolutiontech/scraper-go/internal/pkg/orchestrator/tasks"
	"github.com/exolutiontech/scraper-go/internal/pkg/scrapers/site"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
)

type OrchestratorAPI struct {
	TaskCreator tasks.TaskCreator
	FileStorage storage.FileStorage
	BqClient    bigquery.Client
}

func New(
	taskCreator tasks.TaskCreator,
	fs storage.FileStorage,
	bqc bigquery.Client,
) *OrchestratorAPI {

	return &OrchestratorAPI{
		TaskCreator: taskCreator,
		FileStorage: fs,
		BqClient:    bqc,
	}
}

func (o *OrchestratorAPI) Meta(w http.ResponseWriter, r *http.Request) {
	latestFile, err := o.FileStorage.GetLatestFile("root/", "sites.txt")
	if err != nil {
		errorStr := fmt.Sprintf("file storage error -> GetLatestFile, err: %v", err)
		fmt.Println(errorStr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errorStr))
	}
	fmt.Println("GetLatestFile: ", latestFile)
	urls, err := o.FileStorage.ReadData(latestFile)
	if err != nil {
		errorStr := fmt.Sprintf("file storage error -> ReadData, err: %v", err)
		fmt.Println(errorStr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errorStr))
	}
	fmt.Println("Urls: ", urls)
	if urlStringList, ok := urls.([]string); ok {
		for _, url := range urlStringList {
			err := o.TaskCreator.CreateTaskScrapeMeta(url)
			if err != nil {
				fmt.Println("file storage error: ", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created scrape meta tasks"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("string list not ok"))
	}
}

func (o *OrchestratorAPI) Site(w http.ResponseWriter, r *http.Request) {
	latestFile, err := o.FileStorage.GetLatestFile("root/", "sites.txt")
	if err != nil {
		fmt.Println("file storage error -> GetLatestFile")
	}
	fmt.Println(latestFile)
	urls, err := o.FileStorage.ReadData(latestFile)
	if err != nil {
		fmt.Println("file storage error -> ReadData")
	}
	fmt.Println(urls)
	if urlStringList, ok := urls.([]string); ok {
		for _, url := range urlStringList {
			err := o.TaskCreator.CreateTaskScrapeSite(url)
			if err != nil {
				fmt.Println("file storage error: ", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created scrape site tasks"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("string list not ok"))
	}
}

func (o *OrchestratorAPI) AllMetaProducts(w http.ResponseWriter, r *http.Request) {
	productsFilesPerSite, err := o.FileStorage.GetLatestFiles("meta/", "products.txt")
	if err != nil {
		http.Error(w, "Failed to get products.txt", http.StatusInternalServerError)
		fmt.Println("Error Allproducts: ", err)
		return
	}
	for _, productsFilePerSite := range productsFilesPerSite {
		pFile := strings.ReplaceAll(productsFilePerSite, "meta", "")
		pFile = strings.ReplaceAll(pFile, "/", "")
		err := o.TaskCreator.CreateTaskOrchestrateMetaProduct(pFile)
		if err != nil {
			http.Error(w, "Failed to create product task", http.StatusInternalServerError)
			fmt.Println("Error Allproducts: ", err)
			return
		}
	}
	w.Write([]byte("hopefully created orchestrate scrape meta product"))
}

func (o *OrchestratorAPI) AllProducts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := o.BqClient.Query("SELECT DISTINCT siteIdentifier FROM `scrapers.site_unique`")
	it, err := query.Read(ctx)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		fmt.Println("Error AllMetaProducts: ", err)
		return
	}

	type row struct {
		SiteIdentifier string
	}

	for {
		var values row
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Failed to iterate over results", http.StatusInternalServerError)
			fmt.Println("Error AllProducts: ", err)
			return
		}

		err = o.TaskCreator.CreateTaskOrchestrateProduct(values.SiteIdentifier)
		if err != nil {
			http.Error(w, "Failed to create product task", http.StatusInternalServerError)
			fmt.Println("Error AllProducts: ", err)
			return
		}
	}
	w.Write([]byte("hopefully created orchestrate scrape product"))
}

func (o *OrchestratorAPI) MetaProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	fmt.Println(file)

	startTime := time.Now().UTC()
	// 2 mins to load queue
	startTime = startTime.Add(time.Second * time.Duration(120))
	rateLimit := 1 // Number of requests per second

	splitString := strings.Split(file, "_")
	if len(splitString) == 0 {
		http.Error(w, "Failed to get split string productsFilePerSite", http.StatusInternalServerError)
		return
	}
	siteID := splitString[0]
	siteInfoFile, err := o.FileStorage.GetLatestFile("site/", siteID)
	if err != nil {
		http.Error(w, "Failed to get latest file", http.StatusInternalServerError)
		return
	}

	siteInfoRaw, err := o.FileStorage.ReadData(siteInfoFile)
	if err != nil {
		http.Error(w, "Failed to read site info", http.StatusInternalServerError)
		return
	}
	siteInfoRawT, ok := siteInfoRaw.([]map[string]string)
	if !ok {
		http.Error(w, "Failed to type assert site info raw", http.StatusInternalServerError)
		return
	}
	// todo: do better than this
	var siteInfo []site.SiteData
	for _, siteMap := range siteInfoRawT {
		site := site.SiteData{
			Currency:       siteMap["currency"],
			SiteIdentifier: siteMap["site_identifier"],
			Image:          siteMap["image"],
			Name:           siteMap["name"],
			RateLimit:      siteMap["rate_limit"],
			Scraper:        siteMap["scraper"],
			Technology:     siteMap["technology"],
			Url:            siteMap["url"],
		}
		siteInfo = append(siteInfo, site)
	}
	urlsRaw, err := o.FileStorage.ReadData("meta/" + file)
	if err != nil {
		http.Error(w, "Failed to read product URLs", http.StatusInternalServerError)
		return
	}
	urls, ok := urlsRaw.([]string)
	if !ok {
		http.Error(w, "Failed to type assert site urls raw", http.StatusInternalServerError)
		return
	}
	scheduledTime := startTime
	for _, url := range urls {
		fmt.Println("Scrape Product Task: ", siteInfo[0].Scraper, url)
		scheduledTime = scheduledTime.Add(time.Second * time.Duration(rateLimit))
		err := o.TaskCreator.CreateTaskScrapeProduct(url, siteInfo[0].Scraper, scheduledTime)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to create product task", http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("hopefully created meta product scrape task"))
}
func (o *OrchestratorAPI) Product(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteIdentifier := vars["siteIdentifier"]
	fmt.Println(siteIdentifier)

	startTime := time.Now().UTC()
	// 2 mins to load queue
	startTime = startTime.Add(time.Second * time.Duration(120))
	rateLimit := 1 // Number of requests per second

	ctx := context.Background()

	query := o.BqClient.Query(
		fmt.Sprintf("SELECT * FROM `scrapers.site_unique` WHERE SiteIdentifier = '%s'",
			siteIdentifier))

	it, err := query.Read(ctx)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		fmt.Println("Error MetaProduct: ", err)
		return
	}

	type SiteData struct {
		SiteIdentifier string              `bigquery:"siteidentifier"`
		Name           bigquery.NullString `bigquery:"name"`
		Image          bigquery.NullString `bigquery:"image"`
		Currency       bigquery.NullString `bigquery:"currency"`
		Technology     bigquery.NullString `bigquery:"technology"`
		Scraper        bigquery.NullString `bigquery:"scraper"`
		RateLimit      bigquery.NullString `bigquery:"ratelimit"`
		Url            bigquery.NullString `bigquery:"url"`
	}

	var siteInfo SiteData
	err = it.Next(&siteInfo)
	if err != nil {
		if err == iterator.Done {
			http.Error(w, "No site data found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to iterate over results", http.StatusInternalServerError)
			fmt.Println("Error MetaProduct: ", err)
		}
		return
	}

	query = o.BqClient.Query(`
	SELECT 
		URL 
	FROM (
		SELECT 
			p.URL, 
			p.ProductIdentifier, 
			MAX(d.DateCreated) AS MaxDate
		FROM ` + "`scrapers.product_unique`" + ` AS p
		INNER JOIN ` + "`scrapers.datapoint_raw`" + ` AS d 
			ON p.ProductIdentifier = d.ProductIdentifier
		WHERE 
			p.SiteIdentifier = @siteIdentifier
		GROUP BY 
			p.URL, 
			p.ProductIdentifier
	) as latest
	INNER JOIN ` + "`scrapers.datapoint_raw`" + ` AS dr 
		ON latest.ProductIdentifier = dr.ProductIdentifier 
		AND latest.MaxDate = dr.DateCreated
	WHERE  
		dr.MaxQty > 0
	`)

	query.Parameters = []bigquery.QueryParameter{
		{Name: "siteIdentifier", Value: siteIdentifier},
	}

	it, err = query.Read(ctx)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		fmt.Println("Error Product: ", err)
		return
	}

	type URLData struct {
		URL string `bigquery:"URL"`
	}

	var urlData URLData
	urls := []string{}
	for {
		err := it.Next(&urlData)
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Failed to iterate over results", http.StatusInternalServerError)
			fmt.Println("Error Product: ", err)
			return
		}
		urls = append(urls, urlData.URL)
	}

	scheduledTime := startTime
	for _, url := range urls {
		fmt.Println("URL: ", url)
		if siteInfo.Scraper.Valid {
			fmt.Println("Scrape Product Task: ", siteInfo.Scraper.StringVal, url)
		}
		scheduledTime = scheduledTime.Add(time.Second * time.Duration(rateLimit))
		// Create the product scrape task here
	}

	w.Write([]byte("hopefully created product scrape task"))
}
