package site

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Detect(soup *goquery.Document) string {
	keywords := map[string]string{
		"Shopify.theme":       "shopify",
		"prestashop.com":      "prestashop",
		"cdn3.bigcommerce.com": "bigcommerce",
		"varien/js.js":        "magento",
		"woocommerce":         "woocommerce",
	}

	var technology string

	// Check <script> tags
	soup.Find("script").Each(func(i int, script *goquery.Selection) {
		scriptContent := script.Text()
		for keyword, tech := range keywords {
			if strings.Contains(scriptContent, keyword) {
				technology = tech
				break
			}
		}
	})

	// Check <link> tags
	soup.Find("link").Each(func(i int, link *goquery.Selection) {
		linkContent := link.Text()
		for keyword, tech := range keywords {
			if strings.Contains(linkContent, keyword) {
				technology = tech
				break
			}
		}
	})

	return technology
}