package orchestrator

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/imranparuk/scraper-go/internal/pkg/orchestrator/tasks"
	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/site"

	"github.com/imranparuk/scraper-go/internal/pkg/storage"
)

type OrchestratorAPI struct {
	TaskCreator tasks.TaskCreator
	FileStorage storage.FileStorage
	TargetUrl   string
}

func New(
	taskCreator tasks.TaskCreator,
	fs storage.FileStorage,
	targetUrl string,
) *OrchestratorAPI {
	return &OrchestratorAPI{
		TaskCreator: taskCreator,
		FileStorage: fs,
		TargetUrl: targetUrl,
	}
}

func (o *OrchestratorAPI) Meta(w http.ResponseWriter, r *http.Request) {
	latestFile, err := o.FileStorage.GetLatestFile("root/", "sites.txt")
	if err != nil {
		fmt.Println("file storage error -> GetLatestFile")
	}
	fmt.Println("GetLatestFile: ", latestFile)
	urls, err := o.FileStorage.ReadData(latestFile)
	if err != nil {
		fmt.Println("file storage error -> ReadData")
	}
	fmt.Println("Urls: ", urls)
	if urlStringList, ok := urls.([]string); ok {
		for _, url := range urlStringList {
			err := o.TaskCreator.CreateTaskScrapeMeta(o.TargetUrl, url)
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
			err := o.TaskCreator.CreateTaskScrapeSite(o.TargetUrl, url)
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

func (o *OrchestratorAPI) AllProducts(w http.ResponseWriter, r *http.Request) {
	productsFilesPerSite, err := o.FileStorage.GetLatestFiles("meta/", "products.txt")
	if err != nil {
		http.Error(w, "Failed to get products.txt", http.StatusInternalServerError)
		fmt.Println("Error Allproducts: ", err)
		return
	}
	for _, productsFilePerSite := range productsFilesPerSite {
		pFile := strings.ReplaceAll(productsFilePerSite, "meta", "")
		pFile = strings.ReplaceAll(pFile, "/", "")
		err := o.TaskCreator.CreateTaskOrchestrateProduct(o.TargetUrl, pFile)
		if err != nil {
			http.Error(w, "Failed to create product task", http.StatusInternalServerError)
			fmt.Println("Error Allproducts: ", err)
			return
		}
	}
	w.Write([]byte("hopefully created orchestrate scrape product"))
}

func (o *OrchestratorAPI) Product(w http.ResponseWriter, r *http.Request) {
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
			Currency:   siteMap["currency"],
			ID:         siteMap["id"],
			Image:      siteMap["image"],
			Name:       siteMap["name"],
			RateLimit:  siteMap["rate_limit"],
			Scraper:    siteMap["scraper"],
			Technology: siteMap["technology"],
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
		err := o.TaskCreator.CreateTaskScrapeProduct(o.TargetUrl, url, siteInfo[0].Scraper, scheduledTime)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to create product task", http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("hopefully created product scrape task"))
}
