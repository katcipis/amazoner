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
	"github.com/katcipis/amazoner/product"
)

type Purchase struct {
	Bought   bool
	Reason   string
	Stock    string
	Price    float64
	Delivery string
	DryRun   bool
}

const throttleTime = time.Second

// Do performs a buy with the given parameters.
func Do(link string, maxPrice uint, email, password, userDataDir string, dryRun bool) (*Purchase, error) {

	client := &http.Client{Timeout: 30 * time.Second}

	// FIXME: We have some get/product parsing logic here that could be
	// placed on the product package.
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

	price, err := product.ParsePrice(doc)
	if err != nil {
		return nil, fmt.Errorf("cant parse product price:\n%v", err)
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

	err = makePurchase(link, email, password, userDataDir, availability, dryRun)
	if err != nil {
		return nil, err
	}

	return &Purchase{
		Bought:   true,
		Stock:    availability,
		Reason:   "All conditions met.",
		Price:    price,
		Delivery: delivery,
		DryRun:   dryRun,
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

func makePurchase(link, email, password, userDataDir, availability string, dryRun bool) error {
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
		err = buyFromSellers(session, dryRun)
	default:
		err = buyNow(session, dryRun)
	}

	time.Sleep(throttleTime)

	session.Delete()
	chromeDriver.Stop()

	if err != nil {
		return err
	}

	return nil
}

func buyFromSellers(session *webdriver.Session, dryRun bool) error {

	buySellersBtn, err := session.FindElement(webdriver.ID, "buybox-see-all-buying-choices")
	if err != nil {
		return err
	}

	if err = buySellersBtn.Click(); err != nil {
		return err
	}

	time.Sleep(throttleTime)

	bestOffer, err := getBestOffer(session)
	if err != nil {
		return err
	}

	addToCartBtn, err := bestOffer.FindElement(webdriver.Name, "submit.addToCart")
	if err != nil {
		return err
	}

	if err = addToCartBtn.Click(); err != nil {
		return err
	}

	time.Sleep(throttleTime)

	session.Url("https://www.amazon.com/gp/cart/view.html")

	time.Sleep(throttleTime)

	checkoutBtn, err := session.FindElement(webdriver.ID, "sc-buy-box-ptc-button")
	if err != nil {
		return err
	}

	if err = checkoutBtn.Click(); err != nil {
		return err
	}

	time.Sleep(throttleTime)

	placeOrderBtn, err := session.FindElement(webdriver.ID, "placeYourOrder")
	if err != nil {
		return err
	}

	if !dryRun {
		if err = placeOrderBtn.Click(); err != nil {
			return err
		}
		time.Sleep(throttleTime)
	}

	return nil
}

func buyNow(session *webdriver.Session, dryRun bool) error {
	buyNowBtn, err := session.FindElement(webdriver.ID, "buy-now-button")
	if err != nil {
		return err
	}

	if err = buyNowBtn.Click(); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	if err = session.FocusOnFrame("turbo-checkout-iframe"); err != nil {
		return err
	}
	placeOrderBtn, err := session.FindElement(webdriver.ID, "turbo-checkout-place-order-button")
	if err != nil {
		return err
	}

	if !dryRun {
		if err = placeOrderBtn.Click(); err != nil {
			return err
		}
		time.Sleep(throttleTime)
	}

	return nil
}

func getBestOffer(session *webdriver.Session) (webdriver.WebElement, error) {
	offers, err := session.FindElements(webdriver.ID, "aod-offer")
	if err != nil {
		return webdriver.WebElement{}, err
	}

	if len(offers) == 0 {
		offers, err = session.FindElements(webdriver.CSS_Selector, "#olpOfferList > div > div > div")
		if err != nil {
			return webdriver.WebElement{}, err
		}
	}

	if len(offers) == 0 {
		return webdriver.WebElement{}, errors.New("could not parse best offer from sellers")
	}

	return offers[0], nil
}
