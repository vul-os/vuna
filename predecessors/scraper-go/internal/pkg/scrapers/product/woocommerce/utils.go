package woocommerce

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func PriceToFloat(price interface{}) (float64, error) {
	switch v := price.(type) {
	case string:
		if strings.Contains(v, "R") {
			priceRange := strings.Split(v, "R")
			var lowerValue float64
			for _, p := range priceRange {
				price, err := stringToFloat(p)
				if err != nil {
					return 0, err
				}
				if lowerValue == 0 || price < lowerValue {
					lowerValue = price
				}
			}
			return lowerValue, nil
		} else if strings.Contains(v, "-") {
			// Handle existing range case
			priceRange := strings.Split(v, "-")
			var sum float64
			for _, p := range priceRange {
				price, err := strconv.ParseFloat(p, 64)
				if err != nil {
					return 0, err
				}
				sum += price
			}
			return sum / float64(len(priceRange)), nil
		} else {
			// Handle single value case
			return stringToFloat(v)
		}
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	}
	return 0, fmt.Errorf("unsupported type: %T", price)
}


func MaxQtyToInt(maxQty interface{}) (int, error) {
	switch v := maxQty.(type) {
	case string:
		return stringToInt(v)
	case int:
		return v, nil
	}
	return 0, fmt.Errorf("unsupported type: %T", maxQty)
}

func stringToFloat(s string) (float64, error) {
	cleanedString := regexp.MustCompile(`[^\d.]`).ReplaceAllString(s, "")
	currencyFloat, err := strconv.ParseFloat(cleanedString, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to float: %s", s)
	}
	return currencyFloat, nil
}

func stringToInt(s string) (int, error) {
	pattern := regexp.MustCompile(`\b(\d+)\b`)
	match := pattern.FindStringSubmatch(s)
	if len(match) > 1 {
		number, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("failed to convert string to int: %s", s)
		}
		return number, nil
	}
	return 0, fmt.Errorf("failed to extract number from string: %s", s)
}
