package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	proxyAddress := "p.webshare.io:80"
	proxyUsername := "qnfhspsk-rotate"
	proxyPassword := "t62qs3cx4b6c"
	targetURL := "https://biltongandbudz.co.za/"

	// Create a proxy URL with authentication credentials
	proxyURL, err := url.Parse("http://" + proxyUsername + ":" + proxyPassword + "@" + proxyAddress)
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return
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
		fmt.Println("Error creating request:", err)
		return
	}

	// Perform the HTTP request
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Print the response body
	fmt.Println(string(body))
}
