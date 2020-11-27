package search

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Result struct {
	URL string
}

// Do performs a search with the given parameters.
// It is possible to have results and an error, which indicates
// a partial result.
func Do(name string, minPrice uint, maxPrice uint) ([]Result, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, "https://www.amazon.com/s", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("k", name)
	q.Add("low-price", itoa(minPrice))
	q.Add("high-price", itoa(maxPrice))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// Lazy ignoring the error here
		resBody, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("main search query failed, unexpected status code %d, body:\n%s\n", res.StatusCode, resBody)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	s := doc.Find(".s-main-slot.s-result-list.s-search-results.sg-row")
	s = s.Find("a")
	results := []Result{}

	s.Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			attrs := n.Attr
			for _, attr := range attrs {
				// WHY: Because I can't find a fucking easier way to get the
				// fucking links (fuck CSS selectors and fuck HTML).
				if attr.Key == "href" {
					results = append(results, Result{URL: attr.Val})
				}
			}
		}
	})

	return results, nil
}

func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
