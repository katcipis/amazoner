package search_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/katcipis/amazoner/search"
)

func TestSearch(t *testing.T) {
	// The best we can do is some sort of property conservation
	// test, so we catch bizarre regressions like returning no results
	// or results with empty name, etc (although we dont check specific
	// products or relevance).

	type Test struct {
		domain     string
		search     string
		minPrice   uint
		maxPrice   uint
		minResults uint
	}

	tests := []Test{
		{
			domain:     "www.amazon.com",
			search:     "nvidia rtx 3070",
			minPrice:   500,
			maxPrice:   1500,
			minResults: 13,
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%s%sMin%dMax%dWant%d",
			test.domain,
			test.search,
			test.minPrice,
			test.maxPrice,
			test.minResults,
		)
		t.Run("Searcher/"+testname, func(t *testing.T) {
			searcher := search.New(time.Minute)
			defer searcher.Requester.Close()
			res, err := searcher.Search(test.domain, test.search, test.minPrice, test.maxPrice)
			if len(res) < int(test.minResults) {
				t.Errorf("got %d results; want %d", len(res), test.minResults)
				t.Errorf("results:%v", res)
				if err != nil {
					t.Errorf("errors:%v", err)
				}
			}

			for i, prod := range res {

				if prod.Name == "" {
					t.Errorf("prod %d missing name on product", i)
				}

				if prod.Price <= 0 {
					t.Errorf("prod %d missing price on product", i)
				}
			}
		})
	}
}
