package services

import (
	"fmt"
	"log"

	"github.com/vit0-9/li-enricher-api/parser"
	"github.com/vit0-9/li-enricher-api/scraper"
	"github.com/vit0-9/li-enricher-api/summarizer"
)

// CompanyService encapsulates the business logic for company data enrichment.
type CompanyService struct{}

// NewCompanyService creates a new CompanyService.
func NewCompanyService() *CompanyService {
	return &CompanyService{}
}

// EnrichCompanyData orchestrates the scraping, parsing, and summarization of company data.
// It returns the final data, the type of scrape performed ("full" or "public"), and an error.
func (s *CompanyService) EnrichCompanyData(slug, sessionCookie string) (interface{}, string, error) {
	url := fmt.Sprintf("https://www.linkedin.com/company/%s", slug)

	htmlContent, err := scraper.FetchHTML(url, sessionCookie)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch HTML: %w", err)
	}

	if sessionCookie != "" {
		// --- FULL SCRAPE LOGIC ---
		log.Println("Service: Performing full scrape.")
		jsonData, err := parser.ExtractCompanyJSON(htmlContent)
		if err != nil {
			return nil, "full", fmt.Errorf("failed to parse detailed JSON (is session cookie valid?): %w", err)
		}

		summary, err := summarizer.CreateSummary(jsonData)
		if err != nil {
			return nil, "full", fmt.Errorf("failed to summarize data: %w", err)
		}
		return summary, "full", nil
	}

	// --- PUBLIC SCRAPE LOGIC ---
	log.Println("Service: Performing public scrape.")
	jsonData, err := parser.ExtractLdJSONData(htmlContent)
	if err != nil {
		return nil, "public", fmt.Errorf("failed to extract public ld+json data: %w", err)
	}
	return jsonData, "public", nil
}
