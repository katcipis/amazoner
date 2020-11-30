package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseProductPrice(doc *goquery.Document) (float64, bool) {

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

func ParseById(doc *goquery.Document, id string) (string, bool) {
	query := fmt.Sprintf("#%s", id)
	s := doc.Find(query)
	s.Find("script").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})

	parsedValue := standardizeSpaces(s.Text())
	if parsedValue == "" {
		return "", false
	}
	return parsedValue, true
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

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
