package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cruise-price-compare/internal/domain"
)

// QuoteParseResult represents the structured result from LLM
type QuoteParseResult struct {
	SailingCode   string        `json:"sailing_code"`
	ShipName      string        `json:"ship_name"`
	DepartureDate string        `json:"departure_date"` // YYYY-MM-DD
	Nights        int           `json:"nights"`
	Route         string        `json:"route"`
	Quotes        []ParsedQuote `json:"quotes"`
}

// ParsedQuote represents a single quote from the parsed result
type ParsedQuote struct {
	CabinTypeName string  `json:"cabin_type_name"`
	CabinCategory string  `json:"cabin_category"` // 内舱/海景/阳台/套房
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	PricingUnit   string  `json:"pricing_unit"` // PER_PERSON/PER_CABIN/TOTAL
	Conditions    string  `json:"conditions"`
	Promotion     string  `json:"promotion"`
	Notes         string  `json:"notes"`
}

// ResponseParser handles parsing of LLM responses
type ResponseParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseQuoteResponse parses LLM response into structured quote data
func (p *ResponseParser) ParseQuoteResponse(llmResponse string) (*QuoteParseResult, error) {
	// Clean the response - LLMs sometimes wrap JSON in markdown code blocks
	cleanedResponse := p.cleanLLMResponse(llmResponse)

	// Try to parse as JSON
	var result QuoteParseResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w. Response: %s", err, cleanedResponse)
	}

	// Validate the parsed result
	if err := p.validateResult(&result); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &result, nil
}

// cleanLLMResponse removes common LLM response artifacts
func (p *ResponseParser) cleanLLMResponse(response string) string {
	// Trim whitespace
	cleaned := strings.TrimSpace(response)

	// Remove markdown code block markers
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}

	// Trim again after removing markers
	cleaned = strings.TrimSpace(cleaned)

	// Sometimes LLMs add explanatory text before/after JSON
	// Try to extract just the JSON part
	if start := strings.Index(cleaned, "{"); start != -1 {
		if end := strings.LastIndex(cleaned, "}"); end != -1 && end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	return cleaned
}

// validateResult validates the parsed result
func (p *ResponseParser) validateResult(result *QuoteParseResult) error {
	// Validate sailing information
	if result.SailingCode == "" {
		return fmt.Errorf("sailing_code is required")
	}
	if result.ShipName == "" {
		return fmt.Errorf("ship_name is required")
	}
	if result.Nights <= 0 {
		return fmt.Errorf("nights must be positive")
	}

	// Validate departure date format
	if result.DepartureDate != "" {
		if _, err := time.Parse("2006-01-02", result.DepartureDate); err != nil {
			return fmt.Errorf("departure_date must be in YYYY-MM-DD format: %w", err)
		}
	}

	// Validate quotes
	if len(result.Quotes) == 0 {
		return fmt.Errorf("at least one quote is required")
	}

	for i, quote := range result.Quotes {
		if err := p.validateQuote(&quote, i); err != nil {
			return fmt.Errorf("quote[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

// validateQuote validates a single quote
func (p *ResponseParser) validateQuote(quote *ParsedQuote, index int) error {
	if quote.CabinTypeName == "" {
		return fmt.Errorf("cabin_type_name is required")
	}

	if quote.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}

	// Validate currency (must be 3-letter code)
	if len(quote.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter code (e.g., USD, CNY)")
	}

	// Validate pricing unit
	validUnits := map[string]bool{
		"PER_PERSON": true,
		"PER_CABIN":  true,
		"TOTAL":      true,
	}
	if !validUnits[quote.PricingUnit] {
		return fmt.Errorf("pricing_unit must be one of: PER_PERSON, PER_CABIN, TOTAL")
	}

	// Validate cabin category if provided
	if quote.CabinCategory != "" {
		validCategories := map[string]bool{
			"内舱": true,
			"海景": true,
			"阳台": true,
			"套房": true,
		}
		if !validCategories[quote.CabinCategory] {
			return fmt.Errorf("cabin_category must be one of: 内舱, 海景, 阳台, 套房")
		}
	}

	return nil
}

// ConvertPricingUnit converts string pricing unit to domain enum
func (p *ResponseParser) ConvertPricingUnit(unit string) domain.PricingUnit {
	switch unit {
	case "PER_PERSON":
		return domain.PricingUnitPerPerson
	case "PER_CABIN":
		return domain.PricingUnitPerCabin
	case "TOTAL":
		return domain.PricingUnitTotal
	default:
		return domain.PricingUnitPerPerson // Default fallback
	}
}

// ExtractSailingInfo extracts sailing information from parse result
func (p *ResponseParser) ExtractSailingInfo(result *QuoteParseResult) map[string]interface{} {
	return map[string]interface{}{
		"sailing_code":   result.SailingCode,
		"ship_name":      result.ShipName,
		"departure_date": result.DepartureDate,
		"nights":         result.Nights,
		"route":          result.Route,
	}
}

// TryRecoverFromError attempts to recover from parsing errors
// This is useful when LLM responses are partially correct
func (p *ResponseParser) TryRecoverFromError(llmResponse string, parseErr error) (*QuoteParseResult, error) {
	// Attempt 1: Try to find and fix common JSON syntax errors
	fixed := p.fixCommonJSONErrors(llmResponse)
	if fixed != llmResponse {
		var result QuoteParseResult
		if err := json.Unmarshal([]byte(fixed), &result); err == nil {
			if validateErr := p.validateResult(&result); validateErr == nil {
				return &result, nil
			}
		}
	}

	// Attempt 2: Try to extract partial data
	// This could involve regex patterns or more sophisticated parsing
	// For now, return the original error
	return nil, fmt.Errorf("recovery failed: %w", parseErr)
}

// fixCommonJSONErrors attempts to fix common JSON formatting issues
func (p *ResponseParser) fixCommonJSONErrors(jsonStr string) string {
	// Remove trailing commas before closing brackets/braces
	fixed := strings.ReplaceAll(jsonStr, ",}", "}")
	fixed = strings.ReplaceAll(fixed, ",]", "]")

	// Fix single quotes to double quotes (common LLM mistake)
	// This is naive and may not work in all cases
	// A proper implementation would use a more sophisticated approach

	return fixed
}
