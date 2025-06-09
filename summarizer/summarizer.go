package summarizer

import (
	"fmt"
	"time"
)

func safeGetString(data map[string]interface{}, path ...string) string {
	var current interface{} = data
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}
	s, _ := current.(string)
	return s
}

// safeGet is a helper to get a nested value without asserting its type.
func safeGet(data map[string]interface{}, path ...string) interface{} {
	var current interface{} = data
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = m[key]
		if !ok {
			return nil
		}
	}
	return current
}

// CreateSummary transforms the raw data map into a structured summary.
func CreateSummary(data map[string]interface{}) (map[string]interface{}, error) {
	// The raw JSON has an 'included' array. We need to find the company object within it.
	included, ok := data["included"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("'included' field is not a valid array")
	}

	var companyData map[string]interface{}
	for _, item := range included {
		if obj, ok := item.(map[string]interface{}); ok {
			if pageType, _ := obj["pageType"].(string); pageType == "COMPANY" {
				companyData = obj
				break
			}
		}
	}

	if companyData == nil {
		return nil, fmt.Errorf("could not find company data object in 'included' array")
	}

	// Build the final summary map using our safe accessors.
	summary := make(map[string]interface{})
	summary["name"] = safeGetString(companyData, "name")
	summary["linkedin_handle"] = safeGetString(companyData, "universalName")
	summary["linkedin_profile_url"] = safeGetString(companyData, "url")
	summary["external_id"] = safeGetString(companyData, "entityUrn")
	summary["website"] = safeGetString(companyData, "websiteUrl")
	summary["tagline"] = safeGetString(companyData, "tagline")
	summary["description"] = safeGetString(companyData, "description")

	if foundedOn, ok := safeGet(companyData, "foundedOn").(map[string]interface{}); ok {
		if year, ok := foundedOn["year"].(float64); ok { // JSON numbers are float64
			summary["founded_year"] = int(year)
		}
	}

	if specialities, ok := companyData["specialities"].([]interface{}); ok {
		summary["specialities"] = specialities
	}

	// Safely extract employee count range
	if empRange, ok := safeGet(companyData, "employeeCountRange").(map[string]interface{}); ok {
		start, startOk := empRange["start"].(float64)
		end, endOk := empRange["end"].(float64)
		if startOk && endOk {
			summary["employee_count_range"] = fmt.Sprintf("%d-%d", int(start), int(end))
		}
	}

	// The rest of the logic can be added here following the same pattern:
	// - Extract logo URL
	// - Extract office locations
	// - Extract funding summary
	// For brevity, these are left as an exercise but would follow the same safeGet/safeGetString pattern.
	summary["headquarters"] = extractHeadquarters(companyData)
	summary["office_locations"] = extractOfficeLocations(companyData)
	summary["funding_summary"] = extractFundingSummary(companyData)

	return summary, nil
}

func extractHeadquarters(companyData map[string]interface{}) map[string]interface{} {
	hqData, ok := safeGet(companyData, "headquarter").(map[string]interface{})
	if !ok {
		return nil
	}
	address, ok := hqData["address"].(map[string]interface{})
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"is_headquarters": true,
		"city":            safeGetString(address, "city"),
		"state":           safeGetString(address, "geographicArea"),
		"country":         safeGetString(address, "country"),
		"postal_code":     safeGetString(address, "postalCode"),
	}
}

func extractOfficeLocations(companyData map[string]interface{}) []map[string]interface{} {
	locations := []map[string]interface{}{}
	groupedLocations, ok := safeGet(companyData, "groupedLocations").([]interface{})
	if !ok {
		return locations
	}

	for _, locGroup := range groupedLocations {
		if lg, ok := locGroup.(map[string]interface{}); ok {
			if locs, ok := lg["locations"].([]interface{}); ok && len(locs) > 0 {
				if locDetail, ok := locs[0].(map[string]interface{}); ok {
					if address, ok := locDetail["address"].(map[string]interface{}); ok {
						office := map[string]interface{}{
							"is_headquarters": locDetail["headquarter"],
							"city":            address["city"],
							"state":           address["geographicArea"],
							"country":         address["country"],
							"postal_code":     address["postalCode"],
						}
						locations = append(locations, office)
					}
				}
			}
		}
	}
	return locations
}

func extractFundingSummary(companyData map[string]interface{}) map[string]interface{} {
	fundingData, ok := safeGet(companyData, "crunchbaseFundingData").(map[string]interface{})
	if !ok {
		return nil
	}

	summary := map[string]interface{}{
		"total_rounds":           fundingData["numberOfFundingRounds"],
		"crunchbase_profile_url": fundingData["organizationUrl"],
		"crunchbase_funding_url": fundingData["fundingRoundsUrl"],
	}

	if updatedAt, ok := fundingData["updatedAt"].(float64); ok {
		summary["data_last_updated_utc"] = time.Unix(int64(updatedAt), 0).UTC().Format(time.RFC3339)
	}

	if lastRound, ok := fundingData["lastFundingRound"].(map[string]interface{}); ok {
		lrSummary := map[string]interface{}{
			"type": lastRound["localizedFundingType"],
		}
		if announcedOn, ok := lastRound["announcedOn"].(map[string]interface{}); ok {
			year, yOk := announcedOn["year"].(float64)
			month, mOk := announcedOn["month"].(float64)
			day, dOk := announcedOn["day"].(float64)
			if yOk && mOk && dOk {
				lrSummary["announced_on"] = fmt.Sprintf("%d-%02d-%02d", int(year), int(month), int(day))
			}
		}
		summary["last_round"] = lrSummary
	}
	return summary
}
