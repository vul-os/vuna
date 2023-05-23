package site

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func Wapa() {

	readFile, err := os.Open("/workspace/scraper-go/crawler/file.txt")

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	readFile.Close()

	myList := []string{}

	for _, s := range fileLines {
		if (strings.Contains(s, "ac")) {
			continue
		}
		resp, err := http.DefaultClient.Get(fmt.Sprintf("https://%s", s))
		if err != nil {
			fmt.Println(err)
			continue
		}
		data, _ := ioutil.ReadAll(resp.Body) // Ignoring error for example

		wappalyzerClient, err := wappalyzer.New()
		fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
		if err != nil {
			fmt.Println(err)
		}
		for _, k := range Keys(fingerprints) {
			if strings.Contains(k, "WooCommerce") {
				fmt.Println(k)
				myList = append(myList, k)
			}
		}
		fmt.Println(myList)
	}
	file, err := os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
 
	if err == nil {
		datawriter := bufio.NewWriter(file)
 
		for _, data := range myList {
			_, _ = datawriter.WriteString(data + "\n")
		}
	 
		datawriter.Flush()
		file.Close()
	}
 

	// Output: map[Acquia Cloud Platform:{} Amazon EC2:{} Apache:{} Cloudflare:{} Drupal:{} PHP:{} Percona:{} React:{} Varnish:{}]
}

func Keys(m map[string]struct{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
