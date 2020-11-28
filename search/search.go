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

type Product struct {
	Name  string
	Price float64 // Yeah representing money as float is not an good idea in general
}

type Result struct {
	Product
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

	addUserAgent(req)

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

	const throttleTime = time.Second

	for _, relurl := range urls {
		absURL := entrypointURL + relurl
		result, err := getResult(client, absURL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, result)
		// Avoid amazon errors by hammering the website
		time.Sleep(throttleTime)
	}

	return results, toErr(errs)
}

func getResult(c *http.Client, url string) (Result, error) {
	productPage, err := getProduct(c, url)
	if err != nil {
		return Result{}, err
	}
	defer productPage.Close()

	prod, err := parseProduct(productPage)
	if err != nil {
		return Result{}, fmt.Errorf("url %q parse error:%v", url, err)
	}
	return Result{
		URL:     url,
		Product: prod,
	}, nil
}

func getProduct(c *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	addUserAgent(req)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf(
			"url %q unexpected status %d; resp body:\n%s",
			url,
			res.StatusCode,
			string(body),
		)
	}
	return res.Body, nil
}

func addUserAgent(req *http.Request) {
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36")
}

func parseProduct(html io.Reader) (Product, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {

		return Product{}, err
	}

	name := strings.TrimSpace(doc.Find("#productTitle").Text())
	if name == "" {
		return Product{}, errors.New("cant parse product name")
	}

	price, ok := parseProductPrice(doc)
	if !ok {
		return Product{}, errors.New("cant parse product price")
	}

	return Product{
		Name:  name,
		Price: price,
	}, nil
}

func parseProductPrice(doc *goquery.Document) (float64, bool) {
	if price, ok := parseMoney(doc.Find("#priceblock_ourprice").Text()); ok {
		return price, true
	}

	if price, ok := parseMoney(doc.Find("#style_name_0_price").Text()); ok {
		return price, true
	}

	if price, ok := parseMoney(doc.Find("#olp-upd-new > span > a > span.a-size-base.a-color-price").Text()); ok {
		return price, true
	}

	// Handling more price parsing options will give us more product options
	return 0, false
}

func parseMoney(s string) (float64, bool) {
	// Yeah using float for money is not great...
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return v, true
	}
	sp := strings.Split(s, "$")
	if len(sp) <= 1 {
		return 0, false
	}

	v, err = strconv.ParseFloat(sp[1], 64)
	if err != nil {
		return 0, false
	}
	return v, true
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

	if len(urls) == 0 {
		return nil, errors.New("unable to find any URLs on search result page")
	}

	urls = removeStartingWith(urls, "s", "x", "gp")
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
