package filter

import (
	"errors"
	"strings"

	"github.com/katcipis/amazoner/search"
)

func Do(name string, maxPrice uint, results []search.Result) (string, error) {
	cheaperPrice := float64(maxPrice)
	cheaperResult := search.Result{}
	terms := strings.Fields(name)

	for _, result := range results {
		product := result.Product
		result.Valid = true
		for _, term := range terms {
			if !strings.Contains(strings.ToLower(product.Name), strings.ToLower(term)) {
				result.Valid = false
				break
			}
		}
		if result.Valid && product.Price <= cheaperPrice {
			cheaperPrice = product.Price
			cheaperResult = result
		}
	}

	if cheaperResult.URL != "" {
		return cheaperResult.URL, nil
	} else {
		return "", errors.New("no valid results found")
	}
}
