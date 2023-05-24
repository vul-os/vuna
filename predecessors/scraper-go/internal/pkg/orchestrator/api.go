package orchestrator

import (
	"fmt"
	"net/http"

	"github.com/imranparuk/scraper-go/internal/pkg/orchestrator/tasks"
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
		fmt.Println("file storage error")
	}
	fmt.Println(latestFile)
	urls, err := o.FileStorage.ReadData(latestFile)
	if err != nil {
		fmt.Println("file storage error")
	}
	fmt.Println(urls)
	if urlStringList, ok := urls.([]string); ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("url strings no string list"))
		for _, url := range urlStringList {
			err := o.TaskCreator.CreateTaskMeta(url)
			if err != nil {
				fmt.Println("file storage error")
			}
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created meta tasks"))
	}
}

func (o *OrchestratorAPI) Site(w http.ResponseWriter, r *http.Request) {
	latestFile, err := o.FileStorage.GetLatestFile("root/", "sites.txt")
	if err != nil {
		fmt.Println("file storage error")
	}
	fmt.Println(latestFile)
	urls, err := o.FileStorage.ReadData(latestFile)
	if err != nil {
		fmt.Println("file storage error")
	}
	fmt.Println(urls)
	if urlStringList, ok := urls.([]string); ok {
		for _, url := range urlStringList {
			err := o.TaskCreator.CreateTaskSite(url)
			if err != nil {
				fmt.Println("file storage error")
			}
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hopefully created meta tasks"))
	}
}

// func (o *OrchestratorAPI) Product(w http.ResponseWriter, r *http.Request) {
// 	startTime := time.Now().UTC()

// 	proxyList := createProxyList()
// 	shuffleProxyList(proxyList)
// 	proxyIterator := iter(proxyList)

// 	productsFilesPerSite := o.StorageUtils.GetLatestFiles("meta/", "products.txt")
// 	for _, productsFilePerSite := range productsFilesPerSite {
// 		siteID := strings.Replace(productsFilePerSite, "meta/", "", 1)
// 		siteID = strings.TrimSpace(siteID)

// 		siteInfoFile := o.StorageUtils.GetLatestFile("site/", siteID)
// 		siteInfo := o.StorageUtils.ReadData(siteInfoFile)

// 		rateLimit := 1 // Number of requests per second

// 		if len(siteInfo) > 0 && len(siteInfo[0]) > 0 {
// 			scraperCodeLoc := fmt.Sprintf("scraper_code%s", string(siteInfo[0][len(siteInfo[0])-1]))
// 			blob := o.StorageUtils.Bucket.Blob(scraperCodeLoc)
// 			scraperCode := blob.DownloadAsText()
// 			if siteInfo != nil {
// 				urls := o.StorageUtils.ReadData(productsFilePerSite)
// 				for _, url := range urls {
// 					url = strings.Replace(url, "https://", "", 1)

// 					proxies := map[string]string{"http": fmt.Sprintf("socks5://%s", next(proxyIterator))}

// 					scheduledTime := startTime.Add(time.Second * time.Duration(rateLimit))
// 					scheduledTimestamp := Timestamp{}
// 					scheduledTimestamp.FromDatetime(scheduledTime)

// 					o.TaskCreator.CreateTaskProduct(url, scraperCode, o.TargetURL, scheduledTimestamp, proxies)
// 				}
// 			}
// 		}
// 	}
// 	return "hopefully created", 200
// }
