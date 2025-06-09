package scraper

import (
	"fmt"
	"net/http"

	"github.com/imroc/req/v3"
)

// FetchHTML sends an HTTP GET request to the specified URL and returns the body as a string.
// It uses a session cookie for authentication and impersonates a browser to avoid being blocked.
func FetchHTML(url, sessionCookie string) (string, error) {
	client := req.C().ImpersonateChrome()

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
	resp, err := r.
		Get(url)

	if err != nil {
		return "", fmt.Errorf("http get request failed: %w", err)
	}

	if !resp.IsSuccessState() {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return resp.String(), nil
}

func ValidateSession(sessionCookie string) (bool, error) {
	client := req.C().ImpersonateChrome()
	// Prevent the client from following redirects. If the cookie is invalid,
	// LinkedIn will respond with a 302/303 redirect to the authwall page.
	// If the cookie is valid, it will respond with a 200 OK.
	client.SetRedirectPolicy(req.NoRedirectPolicy())

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
