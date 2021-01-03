package scrapers

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog/log"
	"net/url"
	"scraper-go/utils"
	"strconv"
	"strings"
)

type varStruct struct {
	VarID      int         `json:"variation_id"`
	Sku		   string	   `json:"sku"`
	MaxQty     interface{} `json:"max_qty"`
	Price      float32     `json:"display_price"`
	Attributes struct {
		Size string `json:"attribute_pa_pack-size"`
	} `json:"attributes"`
}

func Scrape(baseUrl string) {
	storeId := 69
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	robotsTxtUrl := fmt.Sprintf("%s/robots.txt", baseUrl)
	log.Info().Msg(
		fmt.Sprintf(
			"robots.txt URL: %s",
			robotsTxtUrl,
		),
	)
	// Array containing all the known URLs
	knownUrls := []string{}
	// Create a Collector specifically for robots.txt and sitemap urls
	getUrlsCollector := colly.NewCollector()
	productsCollector := colly.NewCollector(
		colly.Async(true),
	)
	productsCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4})

	productsCollector.OnHTML(".summary,.entry-summary,.product-summary,.product-info",
		func(e *colly.HTMLElement) {
			querySelection := e.DOM
			maxQtyReplacer := strings.NewReplacer("in", "", "stock", "")
			priceReplacer := strings.NewReplacer("R", "", ",", "")
			productName := strings.TrimSpace(querySelection.Find(".product_title,.entry-title").Text())
			productMeta := querySelection.Find("div[class='product_meta']")
			type items struct {
				name string
				itemUrl  string
			}
			var tagList []items
			var catList []items
			tags := productMeta.Find(".tagged_as").Children()
			categories := productMeta.Find(".posted_in").Children()
			tags.Each(func(_ int, s *goquery.Selection) {
				itemUrl, exists := s.Attr("href")
				if !(exists) {
					itemUrl = ""
				}
				tagList = append(tagList, items{
					s.Text(),
					itemUrl,
				})
				utils.UpsertItem("tags", "tag", s.Text(), itemUrl, storeId)
			})
			categories.Each(func(_ int, s *goquery.Selection) {
				itemUrl, exists := s.Attr("href")
				if !(exists) {
					itemUrl = ""
				}
				catList = append(catList, items{
					s.Text(),
					itemUrl,
				})
				utils.UpsertItem("categories", "category", s.Text(), itemUrl, storeId)
			})
			utils.UpsertItem("products", "product",
				productName, e.Request.URL.String(), storeId)

			// check for product variations
			variationsString, vsResult := querySelection.
				Find("form[class='variations_form cart']").
				Attr("data-product_variations")

			if vsResult {
				// if there are
				var results []varStruct
				if err := json.Unmarshal([]byte(variationsString), &results); err != nil {
					log.Error().Err(err).Msg("unable to unmarshall product")
					return
				}

				// for each variation...
				for i, result := range results {
					maxQty := utils.MaxQtyIntConverter(result.MaxQty, maxQtyReplacer)
					price := utils.PriceFloatConverter(result.Price, priceReplacer)
					if maxQty > 0 {
						log.Info().Msg(
							fmt.Sprintf(
								`Product Var Data (%d):Name: %s, URL: %s, Price: %f, Qty: %d, VarId: %d, Sku: %s, Attributes: %s, Tags: %s, Categories: %s`,
								i,
								productName,
								e.Request.URL,
								price,
								maxQty,
								result.VarID,
								result.Sku,
								result.Attributes,
								tagList,
								catList,
							),
						)
					}
				}
			} else {
				// otherwise there are no variations
				price := querySelection.Find("span[class*='amount']").Text()
				maxQty := querySelection.Find("p[class*='stock']").Text()
				if strings.TrimSpace(maxQty) == "" {
					maxQtyNew, exists := querySelection.Find("input[class*='qty']").Attr("max")
					if !exists {
						maxQtyNew = "0"
					} else {
						maxQty = maxQtyNew
					}
				}
				priceFloat := utils.PriceFloatConverter(price, priceReplacer)
				maxQtyInt := utils.MaxQtyIntConverter(maxQty, maxQtyReplacer)
				sku := querySelection.Find("span[class*='sku']").Text()
				varIdAddCardButton := querySelection.Find("button[name*='add-to-cart']")
				varIdAddCardButtonValue, exists := varIdAddCardButton.Attr("value")
				if !exists {
					return
				}
				varIdAddCardButtonValueInt, err := strconv.Atoi(varIdAddCardButtonValue)
				if err != nil {
					log.Error().Err(err).Msg("Error Converting VarId String to Int!")
				}
				if maxQtyInt > 0 {
					log.Info().Msg(
						fmt.Sprintf(
							`Product No Var Data :Name: %s, URL: %s, Price: %f, Qty: %d, VarId: %d, Sku: %s, Tags: %s, Categories: %s`,
							productName,
							e.Request.URL,
							priceFloat,
							maxQtyInt,
							varIdAddCardButtonValueInt,
							sku,
							tagList,
							catList,
						),
					)
				}

			}
		},
	)

	// Create a callback on the XPath query searching for the URLs
	getUrlsCollector.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		//knownUrls = append(knownUrls, e.Text)
		if strings.Contains(e.Text, "/product/") {
			knownUrls = append(knownUrls, e.Text)
		}
	})

	getUrlsCollector.OnXML("//sitemapindex/sitemap/loc", func(e *colly.XMLElement) {
		if err := e.Request.Visit(e.Text); err != nil {
			log.Error().Err(err).Msg("unable to visit")
		}
	})
	// visit robots.txt and get the sitemaps
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Err getting hostname",
			),
		)
		panic(err)
	}
	hostName := u.Host
	robotsLines, err := utils.UrlToLines(robotsTxtUrl)
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Err getting hostname",
			),
		)
	}
	for _, line := range robotsLines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "sitemap:") {
			sitemapUrl := strings.Replace(lowerLine, "sitemap: ", "", 1)
			if !(strings.Contains(sitemapUrl, hostName)) {
				sitemapUrl = fmt.Sprintf("%s/%s", baseUrl, strings.TrimRight(sitemapUrl, "/"))
			}
			log.Info().Msg(
				fmt.Sprintf(
					"Sitemap Url: %s",
					sitemapUrl,
				),
			)
			err = getUrlsCollector.Visit(sitemapUrl)
			if err != nil {
				log.Error().Err(err).Msg(
					fmt.Sprintf(
						"Err Sitemap Url",
					),
				)
			}
		}
	}
	// Else visit the known sitemaps
	err = getUrlsCollector.Visit(fmt.Sprintf("%s/%s", baseUrl, "/sitemap_index.xml"))
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Error: sitemap_index.xml",
			),
		)
	}
	err = getUrlsCollector.Visit(fmt.Sprintf("%s/%s", baseUrl, "/sitemap.xml"))
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf(
				"Error: sitemap.xml",
			),
		)
	}
	log.Err(err).Msg(
		fmt.Sprintf(
			"Error: sitemap.xml",
		),
	)
	for _, knownUrl := range knownUrls {
		log.Info().Msg(
			fmt.Sprintf(
				"Visiting Url: %s",
				knownUrl,
			),
		)
		if err := productsCollector.Visit(knownUrl); err != nil {
			log.Error().Err(err).Msg("unable to visit")
		}
	}
	log.Info().Msg(
		fmt.Sprintf(
			"Collected %d URLs",
			len(knownUrls),
		),
	)
	productsCollector.Wait()
}
