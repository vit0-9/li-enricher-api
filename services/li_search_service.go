package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/vit0-9/li-enricher-api/utils"
)

type SearchResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

func SearchCompanies(query, sessionCookie string) ([]SearchResult, error) {
	client := req.C().ImpersonateChrome()
	client.EnableDebugLog()

	csrfToken, jsessionidCookie, err := acquireCsrfToken(sessionCookie, client)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire CSRF token: %w", err)
	}

	apiResponse, err := callSearchAPI(query, csrfToken, sessionCookie, jsessionidCookie, client)
	if err != nil {
		return nil, fmt.Errorf("failed to call LinkedIn search API: %w", err)
	}

	return parseSearchResults(apiResponse)
}

func acquireCsrfToken(sessionCookie string, client *req.Client) (string, *http.Cookie, error) {
	log.Println("Attempting to acquire CSRF token via /feed/")

	resp, err := client.R().
		SetCookies(&http.Cookie{
			Name:  "li_at",
			Value: sessionCookie,
		}).
		Get("https://www.linkedin.com/feed/")
	if err != nil {
		return "", nil, fmt.Errorf("priming request failed: %w", err)
	}
	if !resp.IsSuccessState() {
		return "", nil, fmt.Errorf("priming request returned status: %d", resp.StatusCode)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			csrfToken := strings.Trim(cookie.Value, "\"")
			log.Printf("CSRF token acquired: %s", csrfToken)
			return csrfToken, cookie, nil
		}
	}

	return "", nil, fmt.Errorf("JSESSIONID cookie not found")
}

func callSearchAPI(query, csrfToken, sessionCookie string, jsessionidCookie *http.Cookie, client *req.Client) ([]byte, error) {
	variables := fmt.Sprintf("(query:%s)", query)
	log.Printf("Query variables: %s", variables)

	apiURL := fmt.Sprintf(
		"https://www.linkedin.com/voyager/api/graphql?includeWebMetadata=true&variables=%s&queryId=voyagerSearchDashTypeahead.fa9acbcb761f7b5ec2c808e6da796296",
		variables,
	)

	liAtCookie := &http.Cookie{
		Name:  "li_at",
		Value: sessionCookie,
	}

	log.Printf("Adding cookies: li_at=%s, JSESSIONID=%s", liAtCookie.Value, jsessionidCookie.Value)

	resp, err := client.R().
		SetHeaders(map[string]string{
			"accept":     "application/vnd.linkedin.normalized+json+2.1",
			"csrf-token": csrfToken,
		}).
		SetCookies(liAtCookie, jsessionidCookie).
		Get(apiURL)

	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	if !resp.IsSuccessState() {
		log.Printf("Search API response: %s", resp.String())
		return nil, fmt.Errorf("search request failed: %d, content: %s", resp.StatusCode, resp.String())
	}

	return resp.Bytes(), nil
}

func parseSearchResults(apiResponse []byte) ([]SearchResult, error) {
	var results []SearchResult
	var responseData map[string]interface{}
	if err := json.Unmarshal(apiResponse, &responseData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	elements, ok := utils.SafeGet(responseData, "data", "data", "searchDashTypeaheadByGlobalTypeahead", "elements").([]interface{})
	if !ok {
		return results, nil
	}
	for _, item := range elements {
		elementMap, ok := item.(map[string]interface{})
		if !ok || utils.SafeGetString(elementMap, "suggestionType") != "ENTITY_TYPEAHEAD" {
			continue
		}
		trackingUrn := utils.SafeGetString(elementMap, "entityLockupView", "trackingUrn")
		if strings.HasPrefix(trackingUrn, "urn:li:company:") {
			id := strings.TrimPrefix(trackingUrn, "urn:li:company:")
			results = append(results, SearchResult{
				ID:   id,
				Name: utils.SafeGetString(elementMap, "entityLockupView", "title", "text"),
				Text: utils.SafeGetString(elementMap, "entityLockupView", "subtitle", "text"),
			})
		}
	}
	return results, nil
}
