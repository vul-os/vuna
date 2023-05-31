package product

import (
	"fmt"
	"time"

	"github.com/exolutiontech/scraper-go/internal/pkg/storage"
	"github.com/exolutiontech/scraper-go/internal/pkg/utils"
)

func ToMapAndWriteData(strg storage.FileStorage, list interface{}, fileName string) error {
	dpl, err := utils.ToMap(list)
	fmt.Println(dpl)
	if err != nil {
		return err
	}

	if strg != nil && len(dpl) > 0 {
		err = strg.WriteData(dpl, fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func Save(dataPointList []DataPoint, productDataList []ProductData,
	strg storage.FileStorage, url string, fullScrape bool) error {

	_, encodedSite, err := utils.UrlToIdetifier(url)
	if err != nil {
		return err
	}

	currentDatetime := time.Now()
	formattedDatetime := currentDatetime.Format("2006-01-02-15-04-05")

	err = ToMapAndWriteData(strg, dataPointList,
		fmt.Sprintf("datapoint/%s_%s_dp.csv", encodedSite, formattedDatetime),
	)
	if err != nil {
		return err
	}

	if fullScrape {
		err = ToMapAndWriteData(strg, productDataList,
			fmt.Sprintf("product/%s_%s_product.csv", encodedSite, formattedDatetime),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
