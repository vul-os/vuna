package utils

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type TableRow struct {
	Qty   int
	Price float64
}

type StringReplace struct {
	Str     string
	Replace string
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func MaxQtyIntConverter(qty interface{}, replacer *strings.Replacer) int {
	maxQty := 0
	switch typedResult := qty.(type) {
	case string:
		replacedStr := strings.TrimSpace(replacer.Replace(typedResult))
		i, err := strconv.Atoi(replacedStr)
		if err != nil {
			//log.Info().Msg(
			//	fmt.Sprintf(
			//		"invalid maxqty int string -> Original: %s, Replaced: %s, Int: %d",
			//		strings.TrimSpace(typedResult),
			//		replacedStr,
			//		0,
			//	),
			//)
			return 0
		}
		maxQty = i
	case int:
		maxQty = typedResult
	case float32:
		maxQty = int(typedResult)
	case float64:
		maxQty = int(typedResult)
	default:
		log.Error().Msg("invalid type")
		return 0
	}
	return maxQty
}

func PriceFloatConverter(price interface{}, replacer *strings.Replacer) float32 {
	priceFloat := float32(0)
	switch typedResult := price.(type) {
	case string:
		if priceStr, ok := price.(string); ok {
			replacedStr := strings.TrimSpace(replacer.Replace(priceStr))
			pricef, err := strconv.ParseFloat(replacedStr, 32)
			if err != nil {
				//log.Info().Msg(
				//	fmt.Sprintf(
				//		"invalid price float string -> Original: %s, Replaced: %s, Float: %d",
				//		strings.TrimSpace(typedResult),
				//		replacedStr,
				//		0,
				//	),
				//)
				return 0
			}
			priceFloat = float32(pricef)
		} else {
			log.Error().Msg("invalid price float string -> not a valid string")
			return 0
		}
	case int:
		priceFloat = float32(typedResult)
	case float32:
		priceFloat = typedResult
	case float64:
		priceFloat = float32(typedResult)
	default:
		log.Error().Msg("invalid type")
		priceFloat = 0
	}
	return priceFloat
}
