package proverbs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// LoadAllProverbs returns a complete collection of both official and community proverbs
func LoadAllProverbs() *ProverbCollection {
	// Initialize example loader if not already done
	if globalExampleLoader == nil {
		if err := InitExampleLoader(); err != nil {
			// Log error but continue with empty examples
			fmt.Printf("Warning: Failed to load examples: %v\n", err)
		}
	}
	
	return &ProverbCollection{
		Official:  GetOfficialProverbs(),
		Community: GetCommunityProverbs(),
		UpdatedAt: time.Now(),
	}
}



// LoadFromFile loads proverbs from a JSON file
func LoadFromFile(filename string) (*ProverbCollection, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filename, err)
	}
	
	var collection ProverbCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON: %w", err)
	}
	
	return &collection, nil
}

// SaveToFile saves the proverb collection to a JSON file
func (pc *ProverbCollection) SaveToFile(filename string) error {
	data, err := pc.ToJSON()
	if err != nil {
		return fmt.Errorf("marshaling to JSON: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", filename, err)
	}
	
	return nil
}

// GetStats returns statistics about the proverb collection
func (pc *ProverbCollection) GetStats() ProverbStats {
	stats := ProverbStats{
		Total:     len(pc.Official) + len(pc.Community),
		Official:  len(pc.Official),
		Community: len(pc.Community),
		Categories: make(map[Category]int),
		Tags:       make(map[string]int),
	}
	
	// Count by category and tags
	for _, proverb := range pc.GetAll() {
		stats.Categories[proverb.Category]++
		for _, tag := range proverb.Tags {
			stats.Tags[tag]++
		}
	}
	
	return stats
}

// ProverbStats contains statistics about the proverb collection
type ProverbStats struct {
	Total      int                `json:"total"`
	Official   int                `json:"official"`
	Community  int                `json:"community"`
	Categories map[Category]int   `json:"categories"`
	Tags       map[string]int     `json:"tags"`
}

// SearchProverbs searches for proverbs containing the given text
func (pc *ProverbCollection) SearchProverbs(query string) []Proverb {
	var results []Proverb
	queryLower := strings.ToLower(query)
	
	for _, proverb := range pc.GetAll() {
		if strings.Contains(strings.ToLower(proverb.Title), queryLower) ||
			strings.Contains(strings.ToLower(proverb.Text), queryLower) ||
			strings.Contains(strings.ToLower(proverb.Explanation), queryLower) {
			results = append(results, proverb)
			continue
		}
		
		// Search in tags
		for _, tag := range proverb.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, proverb)
				break
			}
		}
	}
	
	return results
}

// GetRandomProverb returns a random proverb from the collection
func (pc *ProverbCollection) GetRandomProverb() Proverb {
	all := pc.GetAll()
	if len(all) == 0 {
		return Proverb{}
	}
	
	index := rand.Intn(len(all))
	return all[index]
}

// ValidateCollection validates all proverbs in the collection
func (pc *ProverbCollection) ValidateCollection() []ValidationError {
	var errors []ValidationError
	
	// Validate official proverbs
	for id, proverb := range pc.Official {
		errors = append(errors, validateProverb(id, proverb)...)
	}
	
	// Validate community proverbs
	for id, proverb := range pc.Community {
		errors = append(errors, validateProverb(id, proverb)...)
	}
	
	return errors
}

// validateProverb validates a single proverb
func validateProverb(id string, proverb Proverb) []ValidationError {
	var errors []ValidationError
	
	// Validate required fields
	if proverb.Title == "" {
		errors = append(errors, ValidationError{
			ProverbID: id,
			Field:     "Title",
			Message:   "title is required",
		})
	}
	
	if proverb.Text == "" {
		errors = append(errors, ValidationError{
			ProverbID: id,
			Field:     "Text",
			Message:   "text is required",
		})
	}
	
	if proverb.Author == "" {
		errors = append(errors, ValidationError{
			ProverbID: id,
			Field:     "Author",
			Message:   "author is required",
		})
	}
	
	// Validate category
	if !isValidCategory(proverb.Category) {
		errors = append(errors, ValidationError{
			ProverbID: id,
			Field:     "Category",
			Message:   fmt.Sprintf("invalid category: %s", proverb.Category),
		})
	}
	
	// Validate source
	if proverb.Source != SourceOfficial && proverb.Source != SourceCommunity {
		errors = append(errors, ValidationError{
			ProverbID: id,
			Field:     "Source",
			Message:   fmt.Sprintf("invalid source: %s", proverb.Source),
		})
	}
	
	return errors
}

// ValidationError represents a validation error for a proverb
type ValidationError struct {
	ProverbID string `json:"proverb_id"`
	Field     string `json:"field"`
	Message   string `json:"message"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("proverb %s: %s - %s", ve.ProverbID, ve.Field, ve.Message)
}

// isValidCategory checks if the category is valid
func isValidCategory(category Category) bool {
	validCategories := []Category{
		CategorySimplicity,
		CategoryConcurrency,
		CategoryInterfaces,
		CategoryErrors,
		CategoryTesting,
		CategoryPerformance,
		CategoryDesign,
		CategoryIdioms,
		CategoryReflection,
		CategoryPackaging,
	}
	
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}