package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func TestProxies(targetUrl string, proxyList []string, timeout time.Duration) []string {
	if !strings.HasPrefix(targetUrl, "http") {
		targetUrl = "http://" + targetUrl
	}

	var workingProxies []string
	var wg sync.WaitGroup
	proxyChan := make(chan string)

	for _, proxyAddress := range proxyList {
		wg.Add(1)
		go func(proxyAddress string) {
			defer wg.Done()

			proxyURL := &url.URL{
				Scheme: "http",
				Host:   proxyAddress,
			}

			httpTransport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}

			httpClient := &http.Client{
				Transport: httpTransport,
				Timeout:   timeout,
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			request, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
			if err != nil {
				// fmt.Println("Failed to create request for proxy:", proxyAddress)
				return
			}

			response, err := httpClient.Do(request)
			if err != nil {
				// fmt.Println("Request failed for proxy:", proxyAddress)
				fmt.Println(err)
				return
			}

			if response.StatusCode >= 200 && response.StatusCode < 300 {
				proxyChan <- proxyAddress
			}
			response.Body.Close()
		}(proxyAddress)
	}

	// Close the channel after all the goroutines finish
	go func() {
		wg.Wait()
		close(proxyChan)
	}()

	// Collect the working proxies
	for proxyAddress := range proxyChan {
		workingProxies = append(workingProxies, proxyAddress)
	}

	return workingProxies
}
