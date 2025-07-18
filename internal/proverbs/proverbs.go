package proverbs

import (
	"encoding/json"
	"fmt"
	"time"
)

// Proverb represents a single Go proverb with metadata
type Proverb struct {
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	Author      string    `json:"author"`
	Category    Category  `json:"category"`
	Example     string    `json:"example,omitempty"`
	Explanation string    `json:"explanation,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Source      Source    `json:"source"`
}

// Category represents the type of proverb
type Category string

const (
	CategorySimplicity  Category = "simplicity"
	CategoryConcurrency Category = "concurrency"
	CategoryInterfaces  Category = "interfaces"
	CategoryErrors      Category = "errors"
	CategoryTesting     Category = "testing"
	CategoryPerformance Category = "performance"
	CategoryDesign      Category = "design"
	CategoryIdioms      Category = "idioms"
	CategoryReflection  Category = "reflection"
	CategoryPackaging   Category = "packaging"
)

// Source indicates whether the proverb is official or community-contributed
type Source string

const (
	SourceOfficial  Source = "official"
	SourceCommunity Source = "community"
)

// ProverbCollection holds all proverbs organized by source
type ProverbCollection struct {
	Official  map[string]Proverb `json:"official"`
	Community map[string]Proverb `json:"community"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// GetAll returns all proverbs from both sources
func (pc *ProverbCollection) GetAll() []Proverb {
	all := make([]Proverb, 0, len(pc.Official)+len(pc.Community))
	
	// Add official proverbs
	for _, proverb := range pc.Official {
		all = append(all, proverb)
	}
	
	// Add community proverbs
	for _, proverb := range pc.Community {
		all = append(all, proverb)
	}
	
	return all
}

// GetByCategory returns proverbs filtered by category
func (pc *ProverbCollection) GetByCategory(category Category) []Proverb {
	var result []Proverb
	for _, proverb := range pc.GetAll() {
		if proverb.Category == category {
			result = append(result, proverb)
		}
	}
	return result
}

// GetBySource returns proverbs filtered by source
func (pc *ProverbCollection) GetBySource(source Source) []Proverb {
	switch source {
	case SourceOfficial:
		result := make([]Proverb, 0, len(pc.Official))
		for _, proverb := range pc.Official {
			result = append(result, proverb)
		}
		return result
	case SourceCommunity:
		result := make([]Proverb, 0, len(pc.Community))
		for _, proverb := range pc.Community {
			result = append(result, proverb)
		}
		return result
	default:
		return pc.GetAll()
	}
}

// GetByTag returns all proverbs that contain a specific tag
func (pc *ProverbCollection) GetByTag(tag string) []Proverb {
	var result []Proverb
	for _, proverb := range pc.GetAll() {
		for _, proverbTag := range proverb.Tags {
			if proverbTag == tag {
				result = append(result, proverb)
				break
			}
		}
	}
	return result
}

// GetByID returns a proverb by its ID
func (pc *ProverbCollection) GetByID(id string) *Proverb {
	if proverb, exists := pc.Official[id]; exists {
		return &proverb
	}
	if proverb, exists := pc.Community[id]; exists {
		return &proverb
	}
	return nil
}

// ToJSON converts the collection to JSON
func (pc *ProverbCollection) ToJSON() ([]byte, error) {
	return json.MarshalIndent(pc, "", "  ")
}

// String implements the Stringer interface for Category
func (c Category) String() string {
	return string(c)
}

// String implements the Stringer interface for Source
func (s Source) String() string {
	return string(s)
}

// String implements the Stringer interface
func (p Proverb) String() string {
	return fmt.Sprintf("%s: %s (by %s)", p.Title, p.Text, p.Author)
}
