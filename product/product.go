package product

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/katcipis/amazoner/parser"
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

	price, ok := parser.ParseProductPrice(doc)
	if !ok {
		return Product{}, errors.New("cant parse product price")
	}

	return Product{
		Name:  name,
		Price: price,
	}, nil
}
