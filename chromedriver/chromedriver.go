package chromedriver

import (
	"github.com/fedesog/webdriver"
)

func NewSession(entrypointURL string) (*webdriver.ChromeDriver, *webdriver.Session, error) {
	chromeDriver := webdriver.NewChromeDriver("chromedriver")
	err := chromeDriver.Start()
	if err != nil {
		return nil, nil, err
	}

	desired := webdriver.Capabilities{"Platform": "Linux"}
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
