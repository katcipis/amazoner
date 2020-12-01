package header

import "net/http"

func Add(req *http.Request) {
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36")
	// TODO : get authority from the request URL
	req.Header.Add("authority", "www.amazon.com")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("rtt", "100")
	req.Header.Add("downlink", "10")
	req.Header.Add("ect", "4g")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cookie", cookie())
}

func cookie() string {
	return `session-id=138-1486552-1785917; session-id-time=2082787201l; i18n-prefs=USD; sp-cdn="L5Z9:DE"; skin=noskin; ubid-main=133-0026094-6185515; session-token=8s/a5BFqbCmg4b/agSrqTkTYamMzq7VZAdO4muCt7XR3TNZvSYkKJLBczicfYyeydY4xtd+AUJ+wxYuyI+YqmKVxh877jEuRlgBUQPFxl2l8qnvk+VXoLt5yOMyk7kUH8mTkelwiKU1xux3waXmaRi9GzqspHEk9QHSD8Ui/ddfEfSWu7tIx7LVxrflH3sp2vtvJmAnIBAaEwUGI+xEb0EDTDQnrGoDBtADU5sUFgL5/gZyPXOp7E5Z7AyceSlRf; csm-hit=tb:YTW86ZKZQDBF1H4MNT0W+s-YTW86ZKZQDBF1H4MNT0W|1606841550750&t:1606841550750&adb:adblk_no'`
}
