package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
)

func FetchWithProxyList(url string, proxyList []string) ([]byte, error) {
	if proxyList == nil {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		return ioutil.ReadAll(response.Body)
	}

	for _, proxy := range proxyList {
		proxyURL, err := netUrl.Parse(proxy)
		if err != nil {
			fmt.Println("Invalid proxy URL:", proxy)
			continue
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client := &http.Client{Transport: transport}

		response, err := client.Get(url)
		if err != nil {
			fmt.Println("Error requesting URL with proxy:", err)
			continue
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			return body, nil
		}
	}

	return nil, errors.New("Failed to fetch data with any proxies")
}