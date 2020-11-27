package search

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

	urls, err := parseResultsURLs(res.Body)
	if err != nil {
		return nil, err
	}

	var errs []error
	var results []Result

	for _, url := range urls {
		result, err := getResult(url)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, result)
	}

	return results, toErr(errs)
}

func getResult(url string) (Result, error) {
	return Result{
		URL: url,
	}, nil
}

func parseResultsURLs(html io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return nil, err
	}

	s := doc.Find(".s-main-slot.s-result-list.s-search-results.sg-row")
	s = s.Find("a")
	urls := []string{}

	s.Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			attrs := n.Attr
			for _, attr := range attrs {
				if attr.Key == "href" {
					urls = append(urls, attr.Val)
				}
			}
		}
	})

	urls = removeStartingWith(urls, "/s/", "/s?", "/gp/", "javascript")
	urls = removeFragments(urls)
	urls = removeDuplicates(urls)

	return urls, nil
}

func removeFragments(urls []string) []string {
	for i, url := range urls {
		sp := strings.Split(url, "#")
		if len(sp) == 0 {
			continue
		}
		urls[i] = sp[0]
	}
	return urls
}

func removeStartingWith(urls []string, prefixes ...string) []string {
	res := []string{}

	for _, url := range urls {
		hasPrefix := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(url, prefix) {
				hasPrefix = true
				break
			}
		}
		if !hasPrefix {
			res = append(res, url)
		}
	}

	return res
}

func removeDuplicates(urls []string) []string {
	uniq := map[string]struct{}{}

	for _, url := range urls {
		uniq[url] = struct{}{}
	}

	i := 0
	res := make([]string, len(uniq))

	for url, _ := range uniq {
		res[i] = url
		i++
	}

	return res
}

func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

func toErr(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	errmsgs := make([]string, len(errs))
	for i, err := range errs {
		errmsgs[i] = err.Error()
	}
	return errors.New(strings.Join(errmsgs, "\n"))
}
