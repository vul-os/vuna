package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

func TestProxies(url string, proxyList []string, timeout time.Duration) []string {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	var workingProxies []string
	var wg sync.WaitGroup
	proxyChan := make(chan string)

	for _, proxyAddress := range proxyList {
		wg.Add(1)
		go func(proxyAddress string) {
			defer wg.Done()
			if !strings.HasPrefix(proxyAddress, "socks5://") {
				proxyAddress = "socks5://" + proxyAddress
			}

			tbDialer, err := proxy.SOCKS5("tcp", strings.TrimPrefix(proxyAddress, "socks5://"), nil, proxy.Direct)
			if err != nil {
				fmt.Println("Error creating SOCKS5 dialer:", err)
				return
			}

			httpTransport := &http.Transport{
				Dial: tbDialer.Dial,
			}
			httpClient := &http.Client{
				Transport: httpTransport,
				Timeout:   timeout,
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				fmt.Println("Failed to create request for proxy:", proxyAddress)
				return
			}

			response, err := httpClient.Do(request)
			if err != nil {
				fmt.Println("Request failed for proxy:", proxyAddress)
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
