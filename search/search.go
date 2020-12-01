package search

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/katcipis/amazoner/chromedriver"
	// "github.com/katcipis/amazoner/debug"
	"github.com/katcipis/amazoner/header"
	"github.com/katcipis/amazoner/product"
)

// Searcher searches for products and provides the found products
// with an internal cache to avoid getting products details too often.
//
// The searcher is NOT concurrency safe.
type Searcher struct {
	CachePeriod time.Duration
	cache       map[string]cacheEntry
}

type Error string

const (
	ErrCaptcha Error = "captcha challenge"
)

func New(cachePeriod time.Duration) *Searcher {
	return &Searcher{
		CachePeriod: cachePeriod,
		cache:       map[string]cacheEntry{},
	}
}

// Do performs a search with the given parameters and returns
// a list of products. It can produce partial results so you
// should check for the products even if an error is returned.
func (s *Searcher) Search(domain, name string, minPrice, maxPrice uint) ([]product.Product, error) {
	s.cleanCache()
	urls, err := Do(domain, name, minPrice, maxPrice)
	if err != nil {
		return nil, err
	}

	uncachedURLs := []string{}
	products := []product.Product{}

	for _, url := range urls {
		r, ok := s.cache[url]
		if ok {
			products = append(products, r.product)
			continue
		}
		uncachedURLs = append(uncachedURLs, url)
	}

	productsGot, err := product.GetProducts(uncachedURLs)
	s.addCache(productsGot)
	return append(products, productsGot...), err
}

// Do performs a search with the given parameters and returns
// a list of products URLs.
func Do(domain, name string, minPrice, maxPrice uint) ([]string, error) {

	entrypointURL := "https://" + domain

	searchQuery := fmt.Sprintf("%s/s", entrypointURL)
	searchUrl, err := url.Parse(searchQuery)
	if err != nil {
		return nil, err
	}

	q := searchUrl.Query()
	q.Add("k", name)
	q.Add("low-price", itoa(minPrice))
	q.Add("high-price", itoa(maxPrice))
	searchUrl.RawQuery = q.Encode()

	html, err := getHtml(searchUrl.String())
	if err != nil {
		return nil, err
	}

	if isCaptchaChallenge(html) {
		html, err := getHtmlChromedriver(searchUrl.String())
		if err != nil {
			return nil, err
		}

		if isCaptchaChallenge(html) {
			return nil, fmt.Errorf("unable to find product URLs : %w", ErrCaptcha)
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	urls, err := parseResultsURLs(doc)
	if err != nil {
		return nil, err
	}

	for i, relurl := range urls {
		absURL := entrypointURL + relurl
		urls[i] = absURL
	}

	return urls, nil
}

type cacheEntry struct {
	product  product.Product
	deadline time.Time
}

func parseResultsURLs(doc *goquery.Document) ([]string, error) {

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
	// FIXME: Copied from product
	if len(errs) == 0 {
		return nil
	}

	errmsgs := make([]string, len(errs))
	for i, err := range errs {
		errmsgs[i] = err.Error()
	}
	return errors.New(strings.Join(errmsgs, "\n"))
}

func isCaptchaChallenge(html string) bool {
	return strings.Contains(html, "captcha")
}

func (e Error) Error() string {
	return string(e)
}

func (s *Searcher) addCache(prods []product.Product) {
	for _, prod := range prods {
		s.cache[prod.URL] = cacheEntry{
			product:  prod,
			deadline: time.Now().Add(s.CachePeriod),
		}
	}
}

func (s *Searcher) cleanCache() {
	for k, entry := range s.cache {
		if time.Now().After(entry.deadline) {
			delete(s.cache, k)
		}
	}
}

func getHtml(entrypointURL string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, entrypointURL, nil)
	if err != nil {
		return "", err
	}

	header.Add(req)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// body, debugFile, err := debug.Save("search-result-http.html", res.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("error trying to dump html response for debug:%v", err)
	// }
	// defer debugFile.Close()

	// resBody, err := ioutil.ReadAll(body)
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("main search query failed, unexpected status code %d, body:\n%s\n", res.StatusCode, resBody)
	}

	return string(resBody), nil
}

func getHtmlChromedriver(entrypointURL string) (string, error) {
	browser, err := chromedriver.NewBrowser(entrypointURL, "")
	if err != nil {
		return "", err
	}
	defer browser.Close()

	time.Sleep(2 * time.Second)

	html, err := browser.Session.Source()
	if err != nil {
		return "", fmt.Errorf("failed to get html from chromedriver session: %v", err)
	}

	// debugFile, err := debug.SaveString("search-result-chrome.html", html)
	// if err != nil {
	// 	return "", fmt.Errorf("error trying to dump html response for debug:%v", err)
	// }
	// defer debugFile.Close()

	return html, nil
}
