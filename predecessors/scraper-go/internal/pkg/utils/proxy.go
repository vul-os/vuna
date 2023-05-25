package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ProxyConfig struct {
	Address  string
	Username string
	Password string
}

func FetchWithProxy(proxyConfig ProxyConfig, targetURL string) ([]byte, error) {
	// Create a proxy URL with authentication credentials
	proxyURL, err := url.Parse("http://" + proxyConfig.Username + ":" + proxyConfig.Password + "@" + proxyConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy URL: %w", err)
	}

	// Create a new HTTP client with the proxy settings
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// Create a new HTTP request
	request, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Perform the HTTP request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}
