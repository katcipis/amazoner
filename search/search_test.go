package search_test

import (
	"fmt"
	"testing"

	"github.com/katcipis/amazoner/search"
)

func TestSearch(t *testing.T) {
	// The best we can do is some sort of property conservation
	// test, so we catch bizarre regressions.

	type Test struct {
		search     string
		minPrice   uint
		maxPrice   uint
		minResults uint
	}

	tests := []Test{
		{
			search:     "nvidia rtx 3070",
			minPrice:   500,
			maxPrice:   1500,
			minResults: 10,
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%sMin%dMax%dWant%d",
			test.search,
			test.minPrice,
			test.maxPrice,
			test.minResults,
		)
		t.Run(testname, func(t *testing.T) {
			res, err := search.Do(test.search, test.minPrice, test.maxPrice)
			if err != nil {
				t.Error(err)
			}
			if len(res) < int(test.minResults) {
				t.Errorf("got %d results; want %d", len(res), test.minResults)
				t.Errorf("results: %v", res)
			}
		})
	}
}
