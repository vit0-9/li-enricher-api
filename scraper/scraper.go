package scraper

import (
	"fmt"
	"net/http"

	"github.com/imroc/req/v3"
)

// FetchHTML fetches the HTML content of a given URL using a session cookie and an optional proxy.
func FetchHTML(url, sessionCookie, proxyURL string) (string, error) {
	client := req.C().ImpersonateChrome()

	if proxyURL != "" {
		client.SetProxyURL(proxyURL)
	}

	client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	client.SetCommonHeader("Accept-Language", "en-US,en;q=0.9")
	r := client.R()

	if sessionCookie != "" {
		cookie := &http.Cookie{
			Name:  "li_at",
			Value: sessionCookie,
		}
		r.SetCookies(cookie)
	}
	resp, err := r.Get(url)

	if err != nil {
		return "", fmt.Errorf("http get request failed: %w", err)
	}

	if !resp.IsSuccessState() {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return resp.String(), nil
}

func ValidateSession(sessionCookie, proxyURL string) (bool, error) {
	client := req.C().ImpersonateChrome()
	client.SetRedirectPolicy(req.NoRedirectPolicy())
	client.SetCommonHeader("Accept-Language", "en-US,en;q=0.9")

	if proxyURL != "" {
		client.SetProxyURL(proxyURL)
	}

	resp, err := client.R().
		SetCookies(&http.Cookie{
			Name:  "li_at",
			Value: sessionCookie,
		}).
		Get("https://www.linkedin.com/feed/")

	if err != nil {
		return false, fmt.Errorf("request to validation URL failed: %w", err)
	}

	return resp.StatusCode == http.StatusOK, nil
}
