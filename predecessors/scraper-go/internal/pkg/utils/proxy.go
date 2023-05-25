package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/proxy"
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
		}

		response, err := customHTTPGetWithClient(url, httpClient)
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

func customHTTPGet(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func customHTTPGetWithClient(url string, client *http.Client) (*http.Response, error) {
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
