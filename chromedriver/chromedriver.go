package chromedriver

import (
	"github.com/fedesog/webdriver"
)

type Browser struct {
	ChromeDriver *webdriver.ChromeDriver
	Session      *webdriver.Session
}

func NewBrowser(entrypointURL, userDataDir string) (*Browser, error) {
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
	required := webdriver.Capabilities{}

	session, err := chromeDriver.NewSession(desired, required)
	if err != nil {
		return nil, err
	}
	err = session.Url(entrypointURL)
	if err != nil {
		return nil, err
	}

	return &Browser{chromeDriver, session}, nil
}

func (b *Browser) Close() {
	b.Session.Delete()
	b.ChromeDriver.Stop()
}
