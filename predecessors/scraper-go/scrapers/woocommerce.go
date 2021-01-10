package scrapers

import (
	"encoding/json"
	"fmt"
	"errors"
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
	Attributes map[string]string `json:"attributes"`
	AvailabilityHtml string `json:"availability_html"`
}

func Min(values []float32) (min float32, e error) {
	if len(values) == 0 {
		return 0, errors.New("Cannot detect a minimum value in an empty slice")
	}
	min = values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}

	return min, nil
}


func Scrape(baseUrl string, numConcurrency int) {
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	storeId, err := utils.GetStoreIdByUrl(baseUrl)
	if err != nil {
		return
	}
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
	productsCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: numConcurrency})

	productsCollector.OnHTML(".summary,.entry-summary,.product-summary,.product-info",
		func(e *colly.HTMLElement) {
			querySelection := e.DOM
			maxQtyReplacer := strings.NewReplacer("in", "", "stock", "")
			priceReplacer := strings.NewReplacer("R", "", ",", "")
			productName := strings.TrimSpace(querySelection.Find(".product_title,.entry-title").Text())
			productMeta := querySelection.Find("div[class='product_meta']")
			tags := productMeta.Find(".tagged_as").Children()
			categories := productMeta.Find(".posted_in").Children()
			// check for product variations
			variationsString, vsResult := querySelection.
				Find("form[class='variations_form cart']").
				Attr("data-product_variations")

			var products []utils.ProdStruct
			var tagList []utils.Items
			var catList []utils.Items
			tags.Each(func(_ int, s *goquery.Selection) {
				itemUrl, exists := s.Attr("href")
				if !(exists) {
					itemUrl = ""
				}
				tagList = append(tagList, utils.Items{
					s.Text(),
					itemUrl,
				})
			})
			categories.Each(func(_ int, s *goquery.Selection) {
				itemUrl, exists := s.Attr("href")
				if !(exists) {
					itemUrl = ""
				}
				catList = append(catList, utils.Items{
					s.Text(),
					itemUrl,
				})
			})

			if vsResult {
				// if there are
				var results []varStruct
				if err := json.Unmarshal([]byte(variationsString), &results); err != nil {
					log.Error().Err(err).Msg("unable to unmarshall product")
					return
				}

				// for each variation...
				for _, result := range results {
					maxQty := utils.MaxQtyIntConverter(result.MaxQty, maxQtyReplacer)
					if maxQty == 0 {
						maxQtyStr := result.AvailabilityHtml
						r := strings.NewReplacer(
							`<p class="stock in-stock">`, "",
							"</p>", "",
							"in", "",
							"out", "",
							"of", "",
							"stock", "",
						)
						maxQtyStr = r.Replace(maxQtyStr)
						maxQtyStr = strings.TrimSpace(maxQtyStr)
						maxQty, _ = strconv.Atoi(maxQtyStr)
					}
					price := utils.PriceFloatConverter(result.Price, priceReplacer)
					if maxQty > 0 {
						products = append(products, utils.ProdStruct{
							ProductName: productName,
							ProductUrl: e.Request.URL.String(),
							StoreId: storeId,
							Categories: catList,
							Tags: tagList,
							VarID: result.VarID,
							Sku: result.Sku,
							MaxQty: maxQty,
							Price: price,
							Attributes: result.Attributes,
						})
					}
				}
			} else {
				var priceFloat float32 = 0.00
				selector := "p[class*='price'] > span[class*='amount']"
				maxQtySelector := "p[class*='stock in-stock']"
				if strings.Contains(baseUrl, "biltongandbudz") {
					selector = "div[class*='product-info'] > div[class*='price'] > " +
						"p[class*='price'] > span[class*='amount']"
				} else if strings.Contains(baseUrl, "smokinggunseeds") {
					maxQtySelector = "div[class*='avada-availability'] > p[class*='stock in-stock']"
				} else if strings.Contains(baseUrl, "livestainable") {
					maxQtySelector = "span[class*='electro-stock-availability'] > p[class*='stock in-stock']"
					selector = ".single-product-wrapper > span[class='electro-price'] * span[class*='woocommerce-Price-amount amount']"
				}
				if strings.Contains(baseUrl, "livestainable") {
					var prices []float32
					// otherwise there are no variations
					querySelection.Find(selector).Children().Each(func(i int, s *goquery.Selection) {
						prstr := strings.ReplaceAll(strings.ReplaceAll(s.Text(), "R", ""),
							",", "")
						price, err := strconv.ParseFloat(prstr, 64)
						if err != nil {
							log.Error().Err(err).Msg("Conv prices error")
						}
						prices = append(prices, float32(price))
					})
					priceFloat, err = Min(prices)

				} else {
					priceStr := querySelection.Find(selector).Text()
					priceStr = strings.ReplaceAll(strings.ReplaceAll(priceStr, "R", ""), ",", "")
					priceFloat = utils.PriceFloatConverter(priceStr, priceReplacer)
				}

				maxQty := querySelection.Find(maxQtySelector).Text()
				r := strings.NewReplacer(
					"in", "",
					"out", "",
					"of", "",
					"stock", "",
					"(can be backordered)", "",
				)
				maxQty = r.Replace(maxQty)
				if strings.TrimSpace(maxQty) == "" {
					maxQtyNew, exists := querySelection.Find("input[class*='qty']").Attr("max")
					if !exists {
						maxQtyNew = "0"
					} else {
						maxQty = maxQtyNew
					}
				}
				maxQtyInt := utils.MaxQtyIntConverter(maxQty, maxQtyReplacer)
				sku := querySelection.Find("span[class*='sku']").Text()
				sku = strings.ReplaceAll(sku,"SKU:", "")
				sku = strings.ReplaceAll(sku, "sku:", "")
				sku = strings.TrimSpace(sku)
				varIdAddCardButton := querySelection.Find("button[name*='add-to-cart']")
				varIdStr, exists := varIdAddCardButton.Attr("value")
				if !exists {
					return
				}
				varId, err := strconv.Atoi(varIdStr)
				if err != nil {
					log.Error().Err(err).Msg("Error Converting VarId String to Int!")
				}
				if maxQtyInt > 0 && priceFloat > 0 {
					products = append(products, utils.ProdStruct{
						ProductName: productName,
						ProductUrl: e.Request.URL.String(),
						StoreId: storeId,
						Categories: catList,
						Tags: tagList,
						VarID: varId,
						Sku: sku,
						MaxQty: maxQtyInt,
						Price: priceFloat,
						Attributes: nil,
					})
				}

			}
			productId, err := utils.UpsertItem("products", "product",
				productName, e.Request.URL.String(), storeId)
			if err != nil {
				return
			}
			for _, tag := range tagList {
				go utils.UpsertItemAndProductItem("tags", "tag",
					tag.Name, tag.ItemUrl, productId, storeId)
			}
			for _, cat := range catList {
				go utils.UpsertItemAndProductItem("categories", "category",
					cat.Name, cat.ItemUrl, productId, storeId)
			}
			go utils.DoAllDb(
				products,
				productName,
				productId,
				storeId,
				e.Request.URL.String(),
				tagList,
				catList,
			)
		},
	)

	// Create a callback on the XPath query searching for the URLs
	getUrlsCollector.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		//knownUrls = append(knownUrls, e.Text)
		if strings.Contains(e.Text, "/product/") ||
			strings.Contains(e.Text, "/products/") ||
			strings.Contains(e.Text, "bikemarket.co.za/shop/") ||
			strings.Contains(e.Text, "bottic.co.za/buy/")	{
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
	log.Info().Msg("Finished")
}
