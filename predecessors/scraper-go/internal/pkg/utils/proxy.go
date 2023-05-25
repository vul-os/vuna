package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const (
	timeout = 3 * time.Second
)

func FetchWithProxyList(url string, proxyList []string) ([]byte, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	if len(proxyList) == 0 {
		response, err := customHTTPGet(url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		return ioutil.ReadAll(response.Body)
	}

	for _, proxyAddress := range proxyList {
		if !strings.HasPrefix(proxyAddress, "socks5://") {
			proxyAddress = "socks5://" + proxyAddress
		}

		tbDialer, err := proxy.SOCKS5("tcp", strings.TrimPrefix(proxyAddress, "socks5://"), nil, proxy.Direct)
		if err != nil {
			fmt.Println("Error creating SOCKS5 dialer:", err)
			continue
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
			return nil, err
		}

		response, err := httpClient.Do(request)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Println("Request timed out for proxy:", proxyAddress)
			} else if err.Error() == "EOF" {
				fmt.Println("EOF error for proxy:", proxyAddress)
			} 
			continue
		}

		if response != nil {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err == nil && response.StatusCode >= 200 && response.StatusCode < 300 {
				return body, nil
			} else if err != nil {
				return nil, err
			}
		}
	}

	return nil, errors.New("Failed to fetch data with any proxies")
}

func customHTTPGet(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
