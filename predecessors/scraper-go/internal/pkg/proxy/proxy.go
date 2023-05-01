package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

func GetProxyList() ([]string, error) {
	url := "https://spys.me/socks.txt" // replace with the URL of your text file

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	var proxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		proxy := scanner.Text()
		if ip, _, err := net.SplitHostPort(proxy); err == nil {
			if net.ParseIP(ip) != nil {
				proxies = append(proxies, proxy)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	// print the list of proxies
	fmt.Println(proxies)
	return proxies, nil
}
