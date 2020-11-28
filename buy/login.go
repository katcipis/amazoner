package buy

import (
	"github.com/fedesog/webdriver"
)

func Login(session *webdriver.Session, email, password string) error {

	accountList, err := session.FindElement(webdriver.ID, "nav-link-accountList")
	if err != nil {
		return err
	}

	err = accountList.Click()
	if err != nil {
		return err
	}

	emailInput, err := session.FindElement(webdriver.ID, "ap_email")
	if err != nil {
		return err
	}

	err = emailInput.SendKeys(email)
	if err != nil {
		return err
	}

	continueBtn, err := session.FindElement(webdriver.ID, "continue")
	if err != nil {
		return err
	}

	err = continueBtn.Click()
	if err != nil {
		return err
	}

	passwordInput, err := session.FindElement(webdriver.ID, "ap_password")
	if err != nil {
		return err
	}

	err = passwordInput.SendKeys(password)
	if err != nil {
		return err
	}

	signInBtn, err := session.FindElement(webdriver.ID, "signInSubmit")
	if err != nil {
		return err
	}

	err = signInBtn.Click()
	if err != nil {
		return err
	}

	return nil
}
