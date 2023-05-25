package proxy

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func CreateProxyList() ([]string, error) {
	url := "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt"
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	var proxyList []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			proxyList = append(proxyList, line)
		}
	}

	return proxyList, nil
}
