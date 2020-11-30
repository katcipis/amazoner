package chromedriver

import (
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
