package proverbs

import (
	"embed"
	"fmt"
	"strings"
)

// exampleFS will be set from the main package
var exampleFS embed.FS

// SetExampleFS sets the embedded filesystem from the main package
func SetExampleFS(fs embed.FS) {
	exampleFS = fs
}

// ExampleLoader handles loading examples from embedded files
type ExampleLoader struct {
	examples map[string]string
}

// NewExampleLoader creates a new example loader
func NewExampleLoader() (*ExampleLoader, error) {
	loader := &ExampleLoader{
		examples: make(map[string]string),
	}

	if err := loader.loadExamples(); err != nil {
		return nil, fmt.Errorf("loading examples: %w", err)
	}

	return loader, nil
}

// loadExamples reads all .gotmpl files from the embedded filesystem
func (el *ExampleLoader) loadExamples() error {
	// Load official examples
	if err := el.loadExamplesFromDir("internal/proverbs/examples/official"); err != nil {
		return fmt.Errorf("loading official examples: %w", err)
	}
	
	// Load community examples
	if err := el.loadExamplesFromDir("internal/proverbs/examples/community"); err != nil {
		return fmt.Errorf("loading community examples: %w", err)
	}
	
	return nil
}

// loadExamplesFromDir loads examples from a specific directory
func (el *ExampleLoader) loadExamplesFromDir(dir string) error {
	entries, err := exampleFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gotmpl") {
			continue
		}

		filePath := dir + "/" + entry.Name()
		content, err := exampleFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", filePath, err)
		}

		// Extract ID from filename (e.g., "official-001.gotmpl" -> "official-001")
		id := strings.TrimSuffix(entry.Name(), ".gotmpl")
		el.examples[id] = string(content)
	}

	return nil
}

// GetExample returns the example content for a given proverb ID
func (el *ExampleLoader) GetExample(proverbID string) (string, bool) {
	example, exists := el.examples[proverbID]
	return example, exists
}

// GetAllExampleIDs returns all available example IDs
func (el *ExampleLoader) GetAllExampleIDs() []string {
	ids := make([]string, 0, len(el.examples))
	for id := range el.examples {
		ids = append(ids, id)
	}
	return ids
}

// HasExample checks if an example exists for the given proverb ID
func (el *ExampleLoader) HasExample(proverbID string) bool {
	_, exists := el.examples[proverbID]
	return exists
}

// GetExampleStats returns statistics about loaded examples
func (el *ExampleLoader) GetExampleStats() ExampleStats {
	stats := ExampleStats{
		Total:     len(el.examples),
		Official:  0,
		Community: 0,
	}

	for id := range el.examples {
		if strings.HasPrefix(id, "official-") {
			stats.Official++
		} else if strings.HasPrefix(id, "community-") {
			stats.Community++
		}
	}

	return stats
}

// ExampleStats contains statistics about loaded examples
type ExampleStats struct {
	Total     int `json:"total"`
	Official  int `json:"official"`
	Community int `json:"community"`
}

// Global example loader instance
var globalExampleLoader *ExampleLoader

// InitExampleLoader initializes the global example loader
func InitExampleLoader() error {
	loader, err := NewExampleLoader()
	if err != nil {
		return err
	}
	globalExampleLoader = loader
	return nil
}

// GetExampleForProverb returns the example for a proverb, loading from files if available
func GetExampleForProverb(proverbID string) string {
	if globalExampleLoader == nil {
		// Fallback to empty string if loader not initialized
		return ""
	}

	example, exists := globalExampleLoader.GetExample(proverbID)
	if !exists {
		return ""
	}

	return example
}
