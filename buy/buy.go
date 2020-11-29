package buy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedesog/webdriver"
	"github.com/katcipis/amazoner/chromedriver"
	"github.com/katcipis/amazoner/parser"
)

type Purchase struct {
	Bought   bool
	Reason   string
	Stock    string
	Price    float64
	Delivery string
}

const throttleTime = time.Second

// Do performs a buy with the given parameters.
func Do(link string, maxPrice uint, email, password, userDataDir string, dryRun bool) (*Purchase, error) {

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

	availability, ok := parser.ParseById(doc, "availability")
	if !ok {
		return nil, errors.New("could not parse availability due to empty string")
	}

	if !checkAvailability(availability) {
		return &Purchase{
			Bought: false,
			Stock:  availability,
			Reason: "No stock available.",
		}, nil
	}

	price, ok := parser.ParseProductPrice(doc)
	if !ok {
		return nil, errors.New("cant parse product price")
	}

	if !checkPrice(price, maxPrice) {
		return &Purchase{
			Bought: false,
			Stock:  availability,
			Reason: "Price was higher than maximum.",
			Price:  price,
		}, nil
	}

	delivery, ok := parser.ParseById(doc, "deliveryMessageMirId")
	if !ok {
		fmt.Fprintln(os.Stderr, "could not parse delivery due to empty string")
	}

	err = makePurchase(link, email, password, userDataDir, availability)
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

func checkAvailability(availability string) bool {
	outOfStockPhrases := []string{
		"unavailable",
	}
	for _, phrase := range outOfStockPhrases {
		if strings.Contains(availability, phrase) {
			return false
		}
	}
	return true
}

func checkPrice(price float64, maxPrice uint) bool {
	return uint(price) <= maxPrice
}

func makePurchase(link, email, password, userDataDir, availability string) error {
	// Start Chromedriver
	chromeDriver, session, err := chromedriver.NewSession(link, userDataDir)
	if err != nil {
		return err
	}

	time.Sleep(throttleTime)

	if userDataDir == "" {
		err = Login(session, email, password)
		if err != nil {
			return err
		}

		time.Sleep(throttleTime)
	}

	switch availability {
	case "Available from these sellers.":
		err = buyFromSellers(session)
	default:
		err = buyNow(session)
	}

	time.Sleep(throttleTime)

	session.Delete()
	chromeDriver.Stop()

	if err != nil {
		return err
	}

	return nil
}

func buyFromSellers(session *webdriver.Session) error {

	buySellersBtn, err := session.FindElement(webdriver.ID, "buybox-see-all-buying-choices")
	if err != nil {
		return err
	}

	err = buySellersBtn.Click()
	if err != nil {
		return err
	}

	time.Sleep(throttleTime)

	offers, err := session.FindElements(webdriver.ID, "aod-offer")
	if err != nil {
		return err
	}

	if len(offers) == 0 {
		offers, err = session.FindElements(webdriver.CSS_Selector, "#olpOfferList > div > div > div")
		if err != nil {
			return err
		}
	}

	bestOffer := offers[0]

	addToCartBtn, err := bestOffer.FindElement(webdriver.Name, "submit.addToCart")
	if err != nil {
		return err
	}

	err = addToCartBtn.Click()
	if err != nil {
		return err
	}

	time.Sleep(throttleTime)
	checkoutBtn, err := bestOffer.FindElement(webdriver.ID, "nav-cart-count")
	if err != nil {
		return err
	}

	err = checkoutBtn.Click()
	if err != nil {
		return err
	}
	time.Sleep(throttleTime)
	return nil
}

func buyNow(session *webdriver.Session) error {
	buyNowBtn, err := session.FindElement(webdriver.ID, "buy-now-button")
	if err != nil {
		return err
	}

	err = buyNowBtn.Click()
	if err != nil {
		return err
	}

	return nil
}
