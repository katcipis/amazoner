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

func ParsePrice(doc *goquery.Document) (float64, bool) {
	// FIXME: probably just exposing Get or a Parse would be better
	// instead of these very specific parsing functions.

	if price, ok := parseMoney(doc.Find("#price_inside_buybox").Text()); ok {
		return price, true
	}

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

	price, ok := ParsePrice(doc)
	if !ok {
		return Product{}, errors.New("cant parse product price")
	}

	return Product{
		Name:  name,
		Price: price,
	}, nil
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
