package buy

import (
	"github.com/fedesog/webdriver"
)

func Login(chromeDriver *webdriver.ChromeDriver, entrypointURL, email, password string) (*webdriver.Session, error) {
	desired := webdriver.Capabilities{"Platform": "Linux"}
	required := webdriver.Capabilities{}

	session, err := chromeDriver.NewSession(desired, required)
	if err != nil {
		return nil, err
	}
	err = session.Url(entrypointURL)
	if err != nil {
		return nil, err
	}

	accountList, err := session.FindElement(webdriver.ID, "nav-link-accountList")
	if err != nil {
		return nil, err
	}

	err = accountList.Click()
	if err != nil {
		return nil, err
	}

	emailInput, err := session.FindElement(webdriver.ID, "ap_email")
	if err != nil {
		return nil, err
	}

	err = emailInput.SendKeys(email)
	if err != nil {
		return nil, err
	}

	continueBtn, err := session.FindElement(webdriver.ID, "continue")
	if err != nil {
		return nil, err
	}

	err = continueBtn.Click()
	if err != nil {
		return nil, err
	}

	passwordInput, err := session.FindElement(webdriver.ID, "ap_password")
	if err != nil {
		return nil, err
	}

	err = passwordInput.SendKeys(password)
	if err != nil {
		return nil, err
	}

	signInBtn, err := session.FindElement(webdriver.ID, "signInSubmit")
	if err != nil {
		return nil, err
	}

	err = signInBtn.Click()
	if err != nil {
		return nil, err
	}

	return session, nil
}
