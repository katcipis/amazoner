package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/katcipis/amazoner/chromedriver"
	"github.com/katcipis/amazoner/header"
)

type Requester struct {
	Browser *chromedriver.Browser
}

func NewRequester() *Requester {
	return &Requester{}
}

func (r *Requester) Get(url string) (io.Reader, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	header.Add(req)

	c := &http.Client{Timeout: 30 * time.Second}

	const throttleTime = time.Second

	time.Sleep(throttleTime)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if isBotDetected(string(body)) {
		if r.Browser == nil {
			r.Browser, err = chromedriver.NewBrowser("")
			if err != nil {
				return nil, err
			}
		}

		err := r.Browser.Url(url)
		if err != nil {
			return nil, err
		}

		time.Sleep(2 * time.Second)

		html, err := r.Browser.Session.Source()
		if err != nil {
			return nil, fmt.Errorf("failed to get html from chromedriver session: %v", err)
		}

		if isBotDetected(html) {
			return nil, fmt.Errorf("unable to navigate to URL '%s' because bot was detected.", url)
		}

		return strings.NewReader(html), nil
	}

	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf(
			"url %q unexpected status %d; resp body:\n%s",
			url,
			res.StatusCode,
			string(body),
		)
	}

	return res.Body, nil
}

func (r *Requester) Close() {
	if r.Browser != nil {
		r.Browser.Close()
	}
}

func isBotDetected(html string) bool {
	return strings.Contains(html, "captcha") || strings.Contains(html, "automated")
}
