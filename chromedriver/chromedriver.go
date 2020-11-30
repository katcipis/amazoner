package chromedriver

import (
	"errors"
	"strings"
	"time"

	"github.com/fedesog/webdriver"
)

type Driver struct {
	driver  *webdriver.ChromeDriver
	session *webdriver.Session
}

func New(userDataDir string) (*Driver, error) {
	// FIXME: maybe use the Driver
	chromeDriver := webdriver.NewChromeDriver("chromedriver")
	err := chromeDriver.Start()
	if err != nil {
		return nil, err
	}

	desired := webdriver.Capabilities{
		"Platform": "Linux",
	}

	if userDataDir != "" {
		desired["goog:chromeOptions"] = map[string]interface{}{
			"args": []string{"user-data-dir=" + userDataDir},
		}
	}
	session, err := chromeDriver.NewSession(desired, webdriver.Capabilities{})
	if err != nil {
		return nil, err
	}

	return &Driver{
		driver:  chromeDriver,
		session: session,
	}, nil
}

func (d *Driver) Get(url string) error {
	return d.session.Url(url)
}

func (d *Driver) Click(elementID string, timeout time.Duration) error {
	elem, err := d.FindElementByID(elementID, timeout)
	if err != nil {
		return err
	}
	return elem.Click()
}

func (d *Driver) Close() {
	d.session.Delete()
	d.driver.Stop()
}

func NewSession(entrypointURL, userDataDir string) (*webdriver.ChromeDriver, *webdriver.Session, error) {
	// FIXME: maybe use the Driver
	chromeDriver := webdriver.NewChromeDriver("chromedriver")
	err := chromeDriver.Start()
	if err != nil {
		return nil, nil, err
	}

	desired := webdriver.Capabilities{
		"Platform": "Linux",
	}

	if userDataDir != "" {
		desired["goog:chromeOptions"] = map[string]interface{}{
			"args": []string{"user-data-dir=" + userDataDir},
		}
	}
	required := webdriver.Capabilities{}

	session, err := chromeDriver.NewSession(desired, required)
	if err != nil {
		return nil, nil, err
	}
	err = session.Url(entrypointURL)
	if err != nil {
		return nil, nil, err
	}
	return chromeDriver, session, nil
}

func (d *Driver) FindElementByID(id string, timeout time.Duration) (webdriver.WebElement, error) {
	// TODO: Not sure if the lib automatically polls when searching for an element
	// Just to be sure we do polling here. This helps avoid spreading random sleeps
	// everywhere waiting for elements to be rendered.
	const pollingtime = 10 * time.Millisecond

	deadline := time.Now().Add(timeout)
	errs := []error{}

	for time.Now().Before(deadline) {
		elem, err := d.session.FindElement(webdriver.ID, "nav-link-accountList")
		if err != nil {
			errs = append(errs, err)
			time.Sleep(pollingtime)
			continue
		}
		return elem, nil
	}

	return webdriver.WebElement{}, toErr(errs)
}

func toErr(errs []error) error {
	// FIXME: Copied from search and product
	if len(errs) == 0 {
		return nil
	}

	errmsgs := make([]string, len(errs))
	for i, err := range errs {
		errmsgs[i] = err.Error()
	}
	return errors.New(strings.Join(errmsgs, "\n"))
}
