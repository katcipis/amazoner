package product

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
	"github.com/katcipis/amazoner/chromedriver"
)

type Product struct {
	Name  string
	Price float64 // Yeah representing money as float is not an good idea in general
}

func Get(url string) (Product, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Product{}, err
	}
	addUserAgent(req)

	c := &http.Client{Timeout: 30 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return Product{}, err
	}
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return Product{}, fmt.Errorf(
			"url %q unexpected status %d; resp body:\n%s",
			url,
			res.StatusCode,
			string(body),
		)
	}
	return parseProduct(res.Body, url)
}

// ParsePrice will try to parse the product price from the given document.
// If it can't because Javascript execution is required to get the price it
// will use the given url to get the product page and navigate it using
// a headless browser.
func ParsePrice(doc *goquery.Document, url string) (float64, error) {
	// FIXME: probably just exposing Get or a Parse would be better
	// instead of these very specific parsing functions.

	errs := []error{}
	parse := func(cssSelector string) (float64, bool) {
		moneyText := doc.Find(cssSelector).Text()
		if moneyText == "" {
			errs = append(errs, fmt.Errorf("html parsing:selector %q selected nothing", cssSelector))
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
	price, err := navigateAndParseBestBuyingOption(url)
	if err == nil {
		return price, nil
	}

	errs = append(errs, err)
	// Handling more price parsing options will give us more product options
	return 0, toErr(errs)
}

func navigateAndParseBestBuyingOption(url string) (float64, error) {
	const timeout = 10 * time.Second

	driver, err := chromedriver.New("")
	if err != nil {
		return 0, err
	}
	defer driver.Close()

	if err := driver.Get(url); err != nil {
		return 0, err
	}

	if err := driver.Click("buybox-see-all-buying-choices", timeout); err != nil {
		return 0, fmt.Errorf("can't click on See All Buying Options button:%v", err)
	}

	return 0, errors.New("implement this")
}

func addUserAgent(req *http.Request) {
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36")
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
	sp := strings.Split(s, "$")
	if len(sp) <= 1 {
		return 0, fmt.Errorf("can't find currency on %q", s)
	}

	moneyText := strings.Replace(sp[1], ",", "", -1)
	v, err = strconv.ParseFloat(moneyText, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse %q : %v", s, err)
	}
	return v, nil
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
