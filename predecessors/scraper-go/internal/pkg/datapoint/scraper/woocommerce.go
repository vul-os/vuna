package scrapers

import (
	"encoding/json"
	"errors"
	"fmt"
	"scraper-go/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog/log"
)

type varStruct struct {
	VarID            int               `json:"variation_id"`
	Sku              string            `json:"sku"`
	MaxQty           interface{}       `json:"max_qty"`
	Price            float32           `json:"display_price"`
	Attributes       map[string]string `json:"attributes"`
	AvailabilityHtml string            `json:"availability_html"`
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

	// storeId, err := utils.GetStoreIdByUrl(baseUrl)
	// if err != nil {
	// 	return
	// }

	robotsTxtUrl := fmt.Sprintf("%s/robots.txt", baseUrl)
	log.Info().Msg(
		fmt.Sprintf(
			"robots.txt URL: %s",
			robotsTxtUrl,
		),
	)
	productsCollector := colly.NewCollector(
		colly.Async(true),
	)
	productsCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: numConcurrency})

	productsCollector.OnHTML(".summary,.entry-summary,.product-summary,.product-info",
		func(e *colly.HTMLElement) {
			fmt.Println("here")
			querySelection := e.DOM
			maxQtyReplacer := strings.NewReplacer("in", "", "stock", "")
			priceReplacer := strings.NewReplacer("R", "", ",", "")
			productName := strings.TrimSpace(querySelection.Find(".product_title,.entry-title").Text())
			// check for product variations
			variationsString, vsResult := querySelection.
				Find("form[class='variations_form cart']").
				Attr("data-product_variations")

			var products []utils.ProdStruct

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
							ProductUrl:  e.Request.URL.String(),
							StoreId:     5,
							VarID:       result.VarID,
							Sku:         result.Sku,
							MaxQty:      maxQty,
							Price:       price,
							Attributes:  result.Attributes,
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
					// priceFloat, err := Min(prices)

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
				sku = strings.ReplaceAll(sku, "SKU:", "")
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
						ProductUrl:  e.Request.URL.String(),
						StoreId:     5,
						VarID:       varId,
						Sku:         sku,
						MaxQty:      maxQtyInt,
						Price:       priceFloat,
						Attributes:  nil,
					})
				}

			}
			// productId, err := utils.UpsertItem("products", "product",
			// 	productName, e.Request.URL.String(), storeId)
			// if err != nil {
			// 	return
			// }

			// go utils.DoAllDb(
			// 	products,
			// 	productName,
			// 	productId,
			// 	storeId,
			// 	e.Request.URL.String(),
			// 	tagList,
			// 	catList,
			// )
		},
	)

	productsCollector.Wait()
	log.Info().Msg("Finished")
}
