package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LiCompany holds the extracted company data from the LD+JSON block.
// Using a struct provides better type safety and clarity.
type LiCompany struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Website       any    `json:"website,omitempty"`
	Slogan        string `json:"slogan,omitempty"`
	EmployeeCount any    `json:"employee_count,omitempty"`
	Headquarters  string `json:"headquarters,omitempty"`
}

type LdJSON struct {
	Graph []map[string]interface{} `json:"@graph"`
}

// isJSONValid performs a deep check to ensure the parsed JSON contains the required nested keys.
func isJSONValid(data map[string]interface{}) bool {
	// Check for data -> data -> organizationDashCompaniesByUniversalName @TODO: Better way to get the json?
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return false
	}
	nestedDataField, ok := dataField["data"].(map[string]interface{})
	if !ok {
		return false
	}

	// also check for nestedDataField["organizationDashCompaniesByIds"]
	_, ok = nestedDataField["organizationDashCompaniesByUniversalName"]
	if !ok {
		_, ok = nestedDataField["*organizationDashCompaniesByIds"]
		log.Println("Checking for organizationDashCompaniesByUniversalName or organizationDashCompaniesByIds:", ok)
	}
	return ok
}

// ExtractCompanyJSON parses the HTML string, finds all <code> tags with an ID
// starting with "bpr-guid", validates their content, and returns the last valid JSON object.
// This function is intended for pages loaded with a valid session cookie.
func ExtractCompanyJSON(htmlContent string) (map[string]interface{}, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var validResults []map[string]interface{}

	// CSS selector to find all <code> tags where the id attribute starts with "bpr-guid".
	doc.Find(`code[id^="bpr-guid"]`).Each(func(i int, s *goquery.Selection) {
		rawJSON := s.Text()
		if rawJSON == "" {
			return
		}

		var parsedJSON map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &parsedJSON); err != nil {
			return
		}

		if isJSONValid(parsedJSON) {
			validResults = append(validResults, parsedJSON)
		}
	})

	if len(validResults) == 0 {
		return nil, fmt.Errorf("no valid company JSON object found in the HTML")
	}

	if len(validResults) > 1 {
		log.Printf("Found %d valid JSON objects. Returning the last one. Might change in the future", len(validResults))
	}

	log.Printf("✅ Found %d valid JSON objects in the HTML.", len(validResults))

	// Return the last valid result. @TODO: This might change in the future, better solution needed.
	return validResults[len(validResults)-1], nil
}

// ExtractLdJSONData finds and parses the <script type="application/ld+json"> tag
// in public-facing HTML. This is a fallback for when no session cookie is available.
func ExtractLdJSONData(htmlContent string) (*LiCompany, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML for ld+json: %w", err)
	}

	ldJSONScript := doc.Find("script[type='application/ld+json']")
	if ldJSONScript.Length() == 0 {
		return nil, fmt.Errorf("could not find the ld+json script tag in the HTML")
	}

	rawJSON := ldJSONScript.Text()
	var ldData LdJSON
	if err := json.Unmarshal([]byte(rawJSON), &ldData); err != nil {
		return nil, fmt.Errorf("error parsing ld+json data: %w", err)
	}

	if len(ldData.Graph) > 0 {
		for _, item := range ldData.Graph {
			// Check if the item is of type "Organization".
			if itemType, ok := item["@type"].(string); ok && itemType == "Organization" {
				log.Println("✅ Found 'Organization' profile within the ld+json data.")
				LiCompany := &LiCompany{}

				if name, ok := item["name"].(string); ok {
					LiCompany.Name = name
				}
				if description, ok := item["description"].(string); ok {
					LiCompany.Description = description
				}
				if slogan, ok := item["slogan"].(string); ok {
					LiCompany.Slogan = slogan
				}
				if sameAs, ok := item["sameAs"]; ok {
					LiCompany.Website = sameAs
				}

				// Safely extract nested employee count.
				if empInfo, ok := item["numberOfEmployees"].(map[string]interface{}); ok {
					LiCompany.EmployeeCount = empInfo["value"]
				}

				// Safely extract and combine nested address info.
				if addrInfo, ok := item["address"].(map[string]interface{}); ok {
					locality, _ := addrInfo["addressLocality"].(string)
					region, _ := addrInfo["addressRegion"].(string)
					country, _ := addrInfo["addressCountry"].(string)
					parts := []string{}
					if locality != "" {
						parts = append(parts, locality)
					}
					if region != "" {
						parts = append(parts, region)
					}
					if country != "" {
						parts = append(parts, country)
					}
					LiCompany.Headquarters = strings.Join(parts, ", ")
				}

				return LiCompany, nil
			}
		}
	}

	return nil, fmt.Errorf("no 'Organization' profile found in ld+json data")
}
