package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/repo"
)

// DataMatcher handles matching of parsed data to existing database records
type DataMatcher struct {
	shipRepo       *repo.ShipRepository
	sailingRepo    *repo.SailingRepository
	cabinTypeRepo  *repo.CabinTypeRepository
	cruiseLineRepo *repo.CruiseLineRepository
}

// NewDataMatcher creates a new data matcher
func NewDataMatcher(
	shipRepo *repo.ShipRepository,
	sailingRepo *repo.SailingRepository,
	cabinTypeRepo *repo.CabinTypeRepository,
	cruiseLineRepo *repo.CruiseLineRepository,
) *DataMatcher {
	return &DataMatcher{
		shipRepo:       shipRepo,
		sailingRepo:    sailingRepo,
		cabinTypeRepo:  cabinTypeRepo,
		cruiseLineRepo: cruiseLineRepo,
	}
}

// MatchResult represents the result of data matching
type MatchResult struct {
	Sailing    *domain.Sailing
	CabinTypes map[string]*domain.CabinType // Key: parsed cabin type name
	Confidence float64                      // 0.0 to 1.0
	Issues     []string                     // Any issues encountered
}

// MatchSailingData matches parsed sailing data to database records
func (m *DataMatcher) MatchSailingData(ctx context.Context, sailingCode, shipName string, departureDate time.Time, nights int) (*MatchResult, error) {
	result := &MatchResult{
		CabinTypes: make(map[string]*domain.CabinType),
		Confidence: 1.0,
		Issues:     []string{},
	}

	// Step 1: Try to find sailing by exact code match
	sailing, err := m.sailingRepo.GetByCode(ctx, sailingCode)
	if err == nil && sailing != nil {
		// Validate the sailing matches other criteria
		if m.validateSailingMatch(sailing, shipName, departureDate, nights) {
			result.Sailing = sailing
			return result, nil
		}
		result.Issues = append(result.Issues, fmt.Sprintf("Sailing code '%s' found but details don't match", sailingCode))
		result.Confidence -= 0.3
	}

	// Step 2: Try to find ship by name
	ship, err := m.findShipByName(ctx, shipName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for ship: %w", err)
	}
	if ship == nil {
		result.Issues = append(result.Issues, fmt.Sprintf("Ship '%s' not found in database", shipName))
		result.Confidence -= 0.5
		return result, nil
	}

	// Step 3: Try to find sailing by ship + departure date + nights
	// Query sailings for this ship
	sailings, err := m.sailingRepo.ListByShip(ctx, ship.ID)
	if err == nil {
		for _, s := range sailings {
			// Check if dates match (within 1 day tolerance)
			dateDiff := s.DepartureDate.Sub(departureDate).Hours() / 24
			if dateDiff >= -1 && dateDiff <= 1 && s.Nights == nights {
				sailing = &s
				result.Sailing = sailing
				if s.SailingCode != sailingCode {
					result.Issues = append(result.Issues, fmt.Sprintf("Sailing code mismatch: expected '%s', found '%s'", sailingCode, s.SailingCode))
					result.Confidence -= 0.2
				}
				return result, nil
			}
		}
	}

	// Step 4: No exact match found
	result.Issues = append(result.Issues, fmt.Sprintf("No sailing found for ship '%s', departure '%s', nights %d", shipName, departureDate.Format("2006-01-02"), nights))
	result.Confidence = 0.0

	return result, nil
}

// MatchCabinType matches a parsed cabin type name to a database record
func (m *DataMatcher) MatchCabinType(ctx context.Context, shipID uint64, cabinTypeName, cabinCategory string) (*domain.CabinType, float64, error) {
	// Step 1: Get all cabin types for this ship
	cabinTypes, err := m.cabinTypeRepo.ListByShip(ctx, shipID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cabin types: %w", err)
	}

	if len(cabinTypes) == 0 {
		return nil, 0, fmt.Errorf("no cabin types found for ship ID %d", shipID)
	}

	// Step 2: Try exact name match first
	for _, ct := range cabinTypes {
		if strings.EqualFold(ct.Name, cabinTypeName) {
			return &ct, 1.0, nil
		}
	}

	// Step 3: Try fuzzy matching
	bestMatch, bestScore := m.findBestCabinTypeMatch(cabinTypes, cabinTypeName, cabinCategory)
	if bestMatch != nil && bestScore >= 0.6 {
		return bestMatch, bestScore, nil
	}

	// Step 4: No good match found
	return nil, 0, fmt.Errorf("no cabin type match found for '%s'", cabinTypeName)
}

// MatchMultipleCabinTypes matches multiple cabin types at once
func (m *DataMatcher) MatchMultipleCabinTypes(ctx context.Context, shipID uint64, cabinTypeNames []string, categories map[string]string) (map[string]*domain.CabinType, []string, error) {
	matched := make(map[string]*domain.CabinType)
	unmatched := []string{}

	// Get all cabin types for the ship once
	cabinTypes, err := m.cabinTypeRepo.ListByShip(ctx, shipID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cabin types: %w", err)
	}

	// Match each cabin type
	for _, name := range cabinTypeNames {
		category := categories[name]

		// Try exact match first
		var exactMatch *domain.CabinType
		for i := range cabinTypes {
			if strings.EqualFold(cabinTypes[i].Name, name) {
				exactMatch = &cabinTypes[i]
				break
			}
		}

		if exactMatch != nil {
			matched[name] = exactMatch
			continue
		}

		// Try fuzzy match
		bestMatch, bestScore := m.findBestCabinTypeMatch(cabinTypes, name, category)
		if bestMatch != nil && bestScore >= 0.6 {
			matched[name] = bestMatch
		} else {
			unmatched = append(unmatched, name)
		}
	}

	return matched, unmatched, nil
}

// validateSailingMatch checks if a sailing matches the expected criteria
func (m *DataMatcher) validateSailingMatch(sailing *domain.Sailing, shipName string, departureDate time.Time, nights int) bool {
	// Check nights
	if sailing.Nights != nights {
		return false
	}

	// Check departure date (allow 1 day tolerance for timezone issues)
	dateDiff := sailing.DepartureDate.Sub(departureDate).Hours() / 24
	if dateDiff < -1 || dateDiff > 1 {
		return false
	}

	// Ship name check would require loading ship data
	// For now, assume match if code and date/nights are correct

	return true
}

// findShipByName finds a ship by name with fuzzy matching
func (m *DataMatcher) findShipByName(ctx context.Context, shipName string) (*domain.Ship, error) {
	// Try exact match first
	pagination := repo.Pagination{Page: 1, PageSize: 100}
	activestatus := domain.EntityStatusActive
	ships, err := m.shipRepo.List(ctx, pagination, nil, &activestatus)
	if err != nil {
		return nil, err
	}

	normalizedSearch := m.normalizeName(shipName)

	// Exact match
	for _, ship := range ships.Items {
		if strings.EqualFold(ship.Name, shipName) {
			return &ship, nil
		}
	}

	// Fuzzy match
	var bestMatch *domain.Ship
	bestScore := 0.0

	for i := range ships.Items {
		score := m.calculateNameSimilarity(normalizedSearch, m.normalizeName(ships.Items[i].Name))
		if score > bestScore && score >= 0.7 {
			bestScore = score
			bestMatch = &ships.Items[i]
		}
	}

	return bestMatch, nil
}

// findBestCabinTypeMatch finds the best matching cabin type using fuzzy matching
func (m *DataMatcher) findBestCabinTypeMatch(cabinTypes []domain.CabinType, targetName, targetCategory string) (*domain.CabinType, float64) {
	var bestMatch *domain.CabinType
	bestScore := 0.0

	normalizedTarget := m.normalizeName(targetName)

	for i := range cabinTypes {
		score := m.calculateNameSimilarity(normalizedTarget, m.normalizeName(cabinTypes[i].Name))

		// Boost score if category matches
		if targetCategory != "" && cabinTypes[i].Category != nil && cabinTypes[i].Category.Name == targetCategory {
			score += 0.2
			if score > 1.0 {
				score = 1.0
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = &cabinTypes[i]
		}
	}

	return bestMatch, bestScore
}

// normalizeName normalizes a name for comparison
func (m *DataMatcher) normalizeName(name string) string {
	// Convert to lowercase
	normalized := strings.ToLower(name)

	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Remove common prefixes/suffixes
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// calculateNameSimilarity calculates similarity between two names (0.0 to 1.0)
// Uses a simple Levenshtein-like approach
func (m *DataMatcher) calculateNameSimilarity(name1, name2 string) float64 {
	// Quick checks
	if name1 == name2 {
		return 1.0
	}
	if name1 == "" || name2 == "" {
		return 0.0
	}

	// Check if one contains the other
	if strings.Contains(name1, name2) || strings.Contains(name2, name1) {
		shorter := len(name1)
		if len(name2) < shorter {
			shorter = len(name2)
		}
		longer := len(name1)
		if len(name2) > longer {
			longer = len(name2)
		}
		return float64(shorter) / float64(longer)
	}

	// Calculate Levenshtein distance
	distance := m.levenshteinDistance(name1, name2)
	maxLen := len(name1)
	if len(name2) > maxLen {
		maxLen = len(name2)
	}

	return 1.0 - (float64(distance) / float64(maxLen))
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (m *DataMatcher) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
