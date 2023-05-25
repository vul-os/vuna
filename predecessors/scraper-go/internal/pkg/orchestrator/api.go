package orchestrator

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imranparuk/scraper-go/internal/pkg/orchestrator/proxy"
	"github.com/imranparuk/scraper-go/internal/pkg/orchestrator/tasks"
	"github.com/imranparuk/scraper-go/internal/pkg/scrapers/site"
	"github.com/imranparuk/scraper-go/internal/pkg/utils"

	"github.com/imranparuk/scraper-go/internal/pkg/storage"
)

type OrchestratorAPI struct {
	TaskCreator tasks.TaskCreator
	FileStorage storage.FileStorage
	TargetURL   string
}

func New(
	taskCreator tasks.TaskCreator,
	fs storage.FileStorage,
	targetURL string,
) *OrchestratorAPI {
	return &OrchestratorAPI{
		TaskCreator: taskCreator,
		FileStorage: fs,
		TargetURL:   targetURL,
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
			err := o.TaskCreator.CreateTaskMeta(url)
			if err != nil {
				fmt.Println("file storage error: ", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created meta tasks"))
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
			err := o.TaskCreator.CreateTaskSite(url)
			if err != nil {
				fmt.Println("file storage error: ", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created site tasks"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("string list not ok"))
	}
}

func (o *OrchestratorAPI) Product(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().UTC()
	// 2 mins to load queue
	startTime = startTime.Add(time.Second * time.Duration(120))
	rateLimit := 1 // Number of requests per second

	proxyListRaw, err := proxy.CreateProxyList()
	if err != nil {
		http.Error(w, "Failed to get proxy list", http.StatusInternalServerError)
		return
	}
	proxyList := utils.TestProxies("http://biltongandbudz.co.za", proxyListRaw, time.Second * 3)
	if len(proxyList) == 0 {
		http.Error(w, "Failed to get working proxys", http.StatusInternalServerError)
		return
	}
	productsFilesPerSite, err := o.FileStorage.GetLatestFiles("meta/", "products.txt")
	for _, productsFilePerSite := range productsFilesPerSite {
		siteID := strings.Replace(productsFilePerSite, "meta/", "", 1)
		siteID = strings.TrimSpace(siteID)

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
		siteInfo, ok := siteInfoRaw.(site.SiteData)
		if !ok {
			http.Error(w, "Failed to type cast site info", http.StatusInternalServerError)
			return
		}

		urlsRaw, err := o.FileStorage.ReadData(productsFilePerSite)
		if err != nil {
			http.Error(w, "Failed to read product URLs", http.StatusInternalServerError)
			return
		}
		urls, ok := urlsRaw.([]string)
		if err != nil {
			http.Error(w, "Failed to type cast product URLs", http.StatusInternalServerError)
			return
		}

		for _, url := range urls {
			scheduledTime := startTime.Add(time.Second * time.Duration(rateLimit))
			proxies := utils.RandomSample(proxyList, 5)
			err := o.TaskCreator.CreateTaskProduct(url, siteInfo.Scraper, proxies, scheduledTime)
			if err != nil {
				http.Error(w, "Failed to create product task", http.StatusInternalServerError)
				return
			}
		}
	}
	w.Write([]byte("hopefully created"))
}
