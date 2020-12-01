package product

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/katcipis/amazoner/header"
)

type Product struct {
	URL   string
	Name  string
	Price float64 // Yeah representing money as float is not an good idea in general
}

func Get(link string) (Product, error) {
	responseBody, err := doRequest(link)
	if err != nil {
		return Product{}, err
	}
	return parseProduct(responseBody, link)
}

// GetProducts gets all products details from the given URLs.
// It is possible to have results and an error, which indicates
// a partial result.
func GetProducts(urls []string) ([]Product, error) {
	var errs []error
	var prods []Product

	for _, url := range urls {
		product, err := Get(url)
		if err != nil {
			errs = append(errs, fmt.Errorf("url %q : %v", url, err))
			continue
		}
		prods = append(prods, product)
	}

	return prods, toErr(errs)
}

func ParsePrice(doc *goquery.Document, link string) (float64, error) {
	// FIXME: probably just exposing Get or a Parse would be better
	// instead of these very specific parsing functions.

	errs := []error{}
	parse := func(cssSelector string) (float64, bool) {
		moneyText := doc.Find(cssSelector).Text()
		if moneyText == "" {
			errs = append(errs, fmt.Errorf("selector %q selected nothing", cssSelector))
			return 0, false
		}
		price, err := parseMoney(moneyText)
		if err != nil {
			errs = append(errs, err)
			return 0, false
		}
		return price, true
	}

	if price, ok := parse("#price_inside_buybox"); ok {
		return price, nil
	}

	if price, ok := parse("#priceblock_ourprice"); ok {
		return price, nil
	}

	if price, ok := parse("#style_name_0_price"); ok {
		return price, nil
	}

	if price, ok := parse("#olp-upd-new > span > a > span.a-size-base.a-color-price"); ok {
		return price, nil
	}

	if price, ok := parse("#olp-upd-new-used"); ok {
		return price, nil
	}

	if price, ok := parse("#olp-upd-used"); ok {
		return price, nil
	}

	// The easy scrapping parsing didn't work, time to bring the big guns
	price, err := navigateAndParseBestBuyingOption(link)
	if err == nil {
		return price, nil
	}

	errs = append(errs, err)
	// Handling more price parsing options will give us more product options
	return 0, toErr(errs)
}

func Filter(name string, prods []Product) []Product {
	// FIXME: move to product package
	validProds := []Product{}
	terms := strings.Fields(name)

	for _, prod := range prods {
		resultValid := true
		for _, term := range terms {
			if !strings.Contains(strings.ToLower(prod.Name), strings.ToLower(term)) {
				resultValid = false
				break
			}
		}
		if resultValid {
			validProds = append(validProds, prod)
		}

	}

	return validProds
}

func SortByPrice(prods []Product) {
	// FIXME: move to product package
	sort.Slice(prods, func(i, j int) bool {
		return prods[i].Price < prods[j].Price
	})
}

func navigateAndParseBestBuyingOption(link string) (float64, error) {
	linkUrl, err := url.Parse(link)
	if err != nil {
		return 0, err
	}

	productId := filepath.Base(linkUrl.Path)

	entrypointURL := fmt.Sprintf("https://%s/gp/offer-listing/%s", linkUrl.Hostname(), productId)

	responseBody, err := doRequest(entrypointURL)
	if err != nil {
		return 0, err
	}

	doc, err := goquery.NewDocumentFromReader(responseBody)
	if err != nil {
		return 0, err
	}

	cssSelector := "#olpOfferList > div > div > div.a-row.a-spacing-mini.olpOffer > div.a-column.a-span2.olpPriceColumn > span"
	moneyText := doc.Find(cssSelector).Text()
	if moneyText == "" {
		return 0, fmt.Errorf("selector %q selected nothing", cssSelector)
	}
	price, err := parseMoney(moneyText)
	if err != nil {
		return 0, err
	}
	return price, nil

}

func parseProduct(html io.Reader, url string) (Product, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {

		return Product{}, err
	}

	name := strings.TrimSpace(doc.Find("#productTitle").Text())
	if name == "" {
		return Product{}, errors.New("cant parse product name")
	}

	price, err := ParsePrice(doc, url)
	if err != nil {
		return Product{}, fmt.Errorf("cant parse product price:\n%v", err)
	}

	return Product{
		URL:   url,
		Name:  name,
		Price: price,
	}, nil
}

func parseMoney(s string) (float64, error) {
	// Yeah using float for money is not great...
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return v, nil
	}

	re := regexp.MustCompile(`[,.0-9]+`)
	s = re.FindString(s)
	v, err = strconv.ParseFloat(s, 64)
	if err == nil {
		return v, nil
	}

	moneyText := strings.Replace(s, ",", "", -1)
	moneyText = strings.Replace(moneyText, ".", "", -1)
	v, err = strconv.ParseFloat(moneyText, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse %q : %v", s, err)
	}
	// After removing ',' and '.' need to divide by 100 to get cents back
	v = v / 100
	return v, nil
}

func doRequest(link string) (io.Reader, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	header.Add(req)

	c := &http.Client{Timeout: 30 * time.Second}

	const throttleTime = time.Second

	time.Sleep(throttleTime)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf(
			"url %q unexpected status %d; resp body:\n%s",
			link,
			res.StatusCode,
			string(body),
		)
	}
	return res.Body, nil
}
func toErr(errs []error) error {
	// FIXME: Copied from search
	if len(errs) == 0 {
		return nil
	}

	errmsgs := make([]string, len(errs))
	for i, err := range errs {
		errmsgs[i] = err.Error()
	}
	return errors.New(strings.Join(errmsgs, "\n"))
}
