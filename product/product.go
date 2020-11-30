package product

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	return parseProduct(res.Body)
}

func ParsePrice(doc *goquery.Document) (float64, error) {
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

	// Handling more price parsing options will give us more product options
	return 0, toErr(errs)
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

	price, err := ParsePrice(doc)
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
