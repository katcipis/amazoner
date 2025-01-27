package buy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	Stock    string
	Price    float64
	Delivery string
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
		return nil, fmt.Errorf("no stock available: %s", availability)
	}

	price, err := product.ParsePrice(doc, link)
	if err != nil {
		return nil, fmt.Errorf("error parsing the price of product with availability '%s'\n%v", availability, err)
	}

	if uint(price) > maxPrice {
		return nil, fmt.Errorf("could not buy product with availability '%s', price '%v' is higher than maximum '%d'.", availability, price, maxPrice)
	}

	delivery, ok := parser.ParseById(doc, "deliveryMessageMirId")
	if !ok {
		fmt.Fprintln(os.Stderr, "could not parse delivery due to empty string")
	}

	err = makePurchase(link, email, password, userDataDir, availability, dryRun)
	if err != nil {
		return nil, fmt.Errorf("error while making purchase of product with availability '%s', price '%v' and delivery '%s'. err: %v", availability, price, delivery, err)
	}

	return &Purchase{
		Stock:    availability,
		Price:    price,
		Delivery: delivery,
	}, nil
}

func checkAvailability(availability string) bool {
	outOfStockPhrases := []string{
		"niet op voorraad",
		"unavailable",
	}
	for _, phrase := range outOfStockPhrases {
		if strings.Contains(availability, phrase) {
			return false
		}
	}
	return true
}

func makePurchase(link, email, password, userDataDir, availability string, dryRun bool) error {
	// Start Chromedriver
	browser, err := chromedriver.NewBrowser(link, userDataDir)
	if err != nil {
		return err
	}
	defer browser.Close()

	time.Sleep(throttleTime)

	if userDataDir == "" {
		err = Login(browser.Session, email, password)
		if err != nil {
			return err
		}

		time.Sleep(throttleTime)
	}

	switch availability {
	case "Available from these sellers.":
	case "Beschikbaar bij deze verkopers.":
		linkUrl, err := url.Parse(link)
		if err != nil {
			return err
		}

		entrypointURL := "https://" + linkUrl.Hostname()

		err = buyFromSellers(browser.Session, dryRun, entrypointURL)
	default:
		err = buyNow(browser.Session, dryRun)
	}

	time.Sleep(throttleTime)

	if err != nil {
		return err
	}

	return nil
}

func buyFromSellers(session *webdriver.Session, dryRun bool, entrypointURL string) error {

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

	session.Url(entrypointURL + "/gp/cart/view.html")

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

	if dryRun {
		return nil
	}

	if err = placeOrderBtn.Click(); err != nil {
		return err
	}
	time.Sleep(throttleTime)

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

	if dryRun {
		return nil
	}

	if err = placeOrderBtn.Click(); err != nil {
		return err
	}
	time.Sleep(throttleTime)

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
