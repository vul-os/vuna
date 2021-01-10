package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"scraper-go/scrapers"
	"scraper-go/utils"
	"strconv"
	"strings"
)

func main() {
	log.Info().Msg("starting server...")
	http.HandleFunc("/", handler)
	utils.GenerateConnPool()

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error().Err(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	bodyStr := strings.TrimSpace(string(body))
	bodyStrSplit := strings.Split(bodyStr, ",")
	numConcurrency := 4
	if len(bodyStrSplit) > 1 {
		numConc, err := strconv.Atoi(bodyStrSplit[1])
		if err != nil {
			log.Error().Err(err).Msg("numConcurrency is Default {4}")
		} else {
			numConcurrency = numConc
			bodyStr = bodyStrSplit[0]
		}

	}
	baseUrl := strings.TrimSpace(bodyStr)
	storeNameRep := strings.NewReplacer(
		".co.za", "",
		".com", "",
		"https://", "",
		"http://", "",
		"/", "",
	)
	urlRep := strings.NewReplacer(
		"https://", "",
		"http://", "",
		"/", "",
	)
	urlReplaced := urlRep.Replace(baseUrl)
	storeNameReplaced := storeNameRep.Replace(urlReplaced)
	log.Info().Msg(fmt.Sprintf("Recieved Request for store: %s, with url: %s, numConcurrency: %d",
		storeNameReplaced, urlReplaced, numConcurrency))
	utils.UpsertStore(storeNameReplaced, urlReplaced)
	scrapers.Scrape(fmt.Sprintf("https://%s", urlReplaced), numConcurrency)
}

// gcloud builds submit --tag gcr.io/spiderbyte-scapers/scraper-go