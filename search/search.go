package search

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	const (
		domain        = "www.amazon.com"
		entrypointURL = "https://" + domain
	)

	client := &http.Client{Timeout: 30 * time.Second}
	searchQuery := fmt.Sprintf("%s/s", entrypointURL)

	req, err := http.NewRequest(http.MethodGet, searchQuery, nil)
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

	for _, relurl := range urls {
		absURL := entrypointURL + relurl
		result, err := getResult(c, absURL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, result)
	}

	return results, toErr(errs)
}

func getResult(c *http.Client, url string) (Result, error) {
	respBody, err := getProduct(c, url)
	if err != nil {
		return Result{}, err
	}
	defer respBody.Close()

	return Result{
		URL: url,
	}, nil
}

func getProduct(url string) (io.ReadCloser, error) {
	// Funny enough to get product details amazon is
	// more pedantic than with searching, so we need
	// a bunch of headers.
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
					parsedURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					urls = append(urls, parsedURL.Path)
				}
			}
		}
	})

	urls = removeStartingWith(urls, "s", "gp")
	urls = removeReferences(urls)
	urls = removeDuplicates(urls)

	return urls, nil
}

func removeReferences(urls []string) []string {
	newURLs := []string{}

	for _, url := range urls {
		splitedURL := strings.Split(url, "/")
		if len(splitedURL) == 0 {
			continue
		}

		newURL := url
		lastIndex := len(splitedURL) - 1
		lastResource := splitedURL[lastIndex]

		if strings.HasPrefix(lastResource, "ref=") {
			newURL = strings.Join(splitedURL[:lastIndex], "/")
		}

		newURLs = append(newURLs, newURL)
	}

	return newURLs
}

func removeStartingWith(urls []string, resourceNames ...string) []string {
	res := []string{}

	for _, url := range urls {
		splitedURL := strings.Split(url, "/")
		if len(splitedURL) <= 1 {
			continue
		}
		firstResource := splitedURL[1]
		starts := false
		for _, resName := range resourceNames {
			if resName == firstResource {
				starts = true
				break
			}
		}
		if !starts {
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
