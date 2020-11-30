package filter

import (
	"errors"
	"strings"

	"github.com/katcipis/amazoner/search"
)

func Do(name string, results []search.Result) (string, error) {
	cheaperPrice := 10000.0
	cheaperResult := search.Result{}
	terms := strings.Fields(name)

	for _, result := range results {
		product := result.Product
		result.Valid = true
		for _, term := range terms {
			if !strings.Contains(product.Name, term) {
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
