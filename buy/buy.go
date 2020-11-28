package buy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedesog/webdriver"
)

type Purchase struct {
	Bought   bool
	Reason   string
	Stock    string
	Price    float64
	Delivery string
}

// Do performs a buy with the given parameters.
func Do(session *webdriver.Session, link string, maxPrice uint) (*Purchase, error) {

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// Lazy ignoring the error here
		resBody, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("buying request failed, unexpected status code %d, body:\n%s\n", res.StatusCode, resBody)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	availability, err := parseAvailability(doc)
	if err != nil {
		return nil, err
	}

	if !checkAvailability(availability) {
		return &Purchase{
			Bought: false,
			Stock:  availability,
			Reason: "No stock available.",
		}, nil
	}

	price, err := parsePrice(doc)
	if err != nil {
		return nil, err
	}

	if !checkPrice(price, maxPrice) {
		return &Purchase{
			Bought: false,
			Stock:  availability,
			Reason: "Price was higher than maximum.",
			Price:  price,
		}, nil
	}

	delivery, err := parseDelivery(doc)
	if err != nil {
		return nil, err
	}

	err = makePurchase(doc)
	if err != nil {
		return nil, err
	}

	return &Purchase{
		Bought:   true,
		Stock:    availability,
		Reason:   "All conditions met.",
		Price:    price,
		Delivery: delivery,
	}, nil
}

func parseAvailability(doc *goquery.Document) (string, error) {
	return parseById(doc, "availability")
}

func checkAvailability(availability string) bool {
	outOfStockPhrases := []string{
		"Available from these sellers.",
		"unavailable",
	}
	for _, phrase := range outOfStockPhrases {
		if strings.Contains(availability, phrase) {
			return false
		}
	}
	return true
}

func parseDelivery(doc *goquery.Document) (string, error) {
	return parseById(doc, "deliveryMessageMirId")
}

func parsePrice(doc *goquery.Document) (float64, error) {

	price, err := parseById(doc, "price_inside_buybox")
	if err != nil {
		return 0.0, err
	}

	re := regexp.MustCompile(`[0-9]+(\.[0-9]+)?`)
	price = re.FindString(price)
	return strconv.ParseFloat(price, 64)
}

func checkPrice(price float64, maxPrice uint) bool {
	return uint(price) <= maxPrice
}

func makePurchase(doc *goquery.Document) error {
	html, err := doc.Html()
	if err != nil {
		return err
	}
	d1 := []byte(html)
	err = ioutil.WriteFile("/tmp/dat1.html", d1, 0644)

	return err
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func parseById(doc *goquery.Document, id string) (string, error) {
	query := fmt.Sprintf("#%s", id)
	s := doc.Find(query)
	s.Find("script").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})

	parsedValue := standardizeSpaces(s.Text())
	if parsedValue == "" {
		return "", fmt.Errorf("Got empty string when parsing id '%s'.", id)
	}
	return parsedValue, nil
}
