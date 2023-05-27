package site

import (
	"fmt"
	"net/http"
	"strings"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func Detect(resp *http.Response, body []byte) string {
	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fingerprints := wappalyzerClient.Fingerprint(resp.Header, body)
	ecommerceKeywords := map[string]string{
		"Shopify":        "Shopify",
		"WooCommerce":    "WooCommerce",
		"WordPress":      "WooCommerce",
		"BigCommerce":    "BigCommerce",
		"PrestaShop":     "PrestaShop",
		"Magento":        "Magento",
		"OpenCart":       "OpenCart",
		"Volusion":       "Volusion",
		"SquareSpace":    "SquareSpace",
		"Weebly":         "Weebly",
		"Wix":            "Wix",
		"CustomPlatform": "Custom Platform",
		// Add more e-commerce technologies as needed
	}

	for _, k := range Keys(fingerprints) {
		for keyword, tech := range ecommerceKeywords {
			if strings.Contains(k, keyword) {
				return strings.ToLower(tech)
			}
		}
	}

	return ""
}
