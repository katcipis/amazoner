package chromedriver

import (
	"github.com/fedesog/webdriver"
)

func NewSession(entrypointURL, userDataDir string) (*webdriver.ChromeDriver, *webdriver.Session, error) {
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
