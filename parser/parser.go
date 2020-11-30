package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
