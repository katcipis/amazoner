package chromedriver

import (
	"github.com/fedesog/webdriver"
)

type Browser struct {
	ChromeDriver         *webdriver.ChromeDriver
	Session              *webdriver.Session
	DesiredCapabilities  webdriver.Capabilities
	RequiredCapabilities webdriver.Capabilities
}

func NewBrowser(userDataDir string) (*Browser, error) {
	chromeDriver := webdriver.NewChromeDriver("chromedriver")
	err := chromeDriver.Start()
	if err != nil {
		return nil, err
	}

	desiredCapabilities := webdriver.Capabilities{
		"Platform": "Linux",
	}

	chromeArgs := []string{}
	// chromeArgs := []string{
	// 	"headless",
	// }

	if userDataDir != "" {
		chromeArgs = append(chromeArgs, "user-data-dir="+userDataDir)
	}

	desiredCapabilities["goog:chromeOptions"] = map[string]interface{}{
		"args": chromeArgs,
	}

	requiredCapabilities := webdriver.Capabilities{}

	return &Browser{
		chromeDriver,
		nil,
		desiredCapabilities,
		requiredCapabilities,
	}, nil
}

func (b *Browser) Url(entrypointURL string) (err error) {
	if b.Session == nil {
		b.Session, err = b.ChromeDriver.NewSession(b.DesiredCapabilities, b.RequiredCapabilities)
		if err != nil {
			return err
		}
	}

	if err = b.Session.Url(entrypointURL); err != nil {
		return err
	}

	return nil
}

func (b *Browser) Close() {
	b.Session.Delete()
	b.ChromeDriver.Stop()
}
