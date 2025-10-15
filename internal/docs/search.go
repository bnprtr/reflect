package docs

import (
	"sort"
	"strings"

	"github.com/bnprtr/reflect/internal/descriptor"
)

// SearchIndex holds all searchable items for fast lookup.
type SearchIndex struct {
	Items []SearchItem
}

// SearchItem represents a single searchable item.
type SearchItem struct {
	Type     string // "service", "method", "message", "enum"
	Name     string
	FullName string
	Package  string
	Comment  string
	URL      string
}

// SearchResult represents a search result with ranking information.
type SearchResult struct {
	SearchItem
	Score int // Higher score = better match
}

// BuildSearchIndex creates a search index from the registry.
func BuildSearchIndex(reg *descriptor.Registry) *SearchIndex {
	if reg == nil {
		return &SearchIndex{Items: []SearchItem{}}
	}

	var items []SearchItem

	// Index services
	for _, service := range reg.ServicesByName {
		item := SearchItem{
			Type:     "service",
			Name:     string(service.Name()),
			FullName: string(service.FullName()),
			Package:  string(service.ParentFile().Package()),
			Comment:  reg.CommentIndex[string(service.FullName())],
			URL:      "/services/" + string(service.FullName()),
		}
		items = append(items, item)

		// Index methods for this service
		for i := 0; i < service.Methods().Len(); i++ {
			method := service.Methods().Get(i)
			methodName := string(service.FullName()) + "/" + string(method.Name())
			methodItem := SearchItem{
				Type:     "method",
				Name:     string(method.Name()),
				FullName: methodName,
				Package:  string(service.ParentFile().Package()),
				Comment:  reg.CommentIndex[methodName],
				URL:      "/methods/" + methodName,
			}
			items = append(items, methodItem)
		}
	}

	// Index messages
	for _, message := range reg.MessagesByName {
		item := SearchItem{
			Type:     "message",
			Name:     string(message.Name()),
			FullName: string(message.FullName()),
			Package:  string(message.ParentFile().Package()),
			Comment:  reg.CommentIndex[string(message.FullName())],
			URL:      "/types/" + string(message.FullName()),
		}
		items = append(items, item)
	}

	// Index enums
	for _, enum := range reg.EnumsByName {
		item := SearchItem{
			Type:     "enum",
			Name:     string(enum.Name()),
			FullName: string(enum.FullName()),
			Package:  string(enum.ParentFile().Package()),
			Comment:  reg.CommentIndex[string(enum.FullName())],
			URL:      "/types/" + string(enum.FullName()),
		}
		items = append(items, item)
	}

	return &SearchIndex{Items: items}
}

// Search performs a case-insensitive search across the index.
// Returns up to 20 results, ranked by relevance.
func (idx *SearchIndex) Search(query string) []SearchResult {
	if len(query) < 2 {
		return []SearchResult{}
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, item := range idx.Items {
		score := calculateScore(item, query)
		if score > 0 {
			results = append(results, SearchResult{
				SearchItem: item,
				Score:      score,
			})
		}
	}

	// Sort by score (descending), then by type, then by name
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		if results[i].Type != results[j].Type {
			return results[i].Type < results[j].Type
		}
		return results[i].Name < results[j].Name
	})

	// Limit to 20 results
	if len(results) > 20 {
		results = results[:20]
	}

	return results
}

// calculateScore calculates a relevance score for a search item.
// Higher scores indicate better matches.
func calculateScore(item SearchItem, query string) int {
	score := 0
	lowerName := strings.ToLower(item.Name)
	lowerFullName := strings.ToLower(item.FullName)
	lowerComment := strings.ToLower(item.Comment)

	// Exact name match (highest priority)
	if lowerName == query {
		score += 100
	}

	// Name starts with query
	if strings.HasPrefix(lowerName, query) {
		score += 50
	}

	// Name contains query
	if strings.Contains(lowerName, query) {
		score += 25
	}

	// Full name starts with query
	if strings.HasPrefix(lowerFullName, query) {
		score += 40
	}

	// Full name contains query
	if strings.Contains(lowerFullName, query) {
		score += 20
	}

	// Comment contains query
	if strings.Contains(lowerComment, query) {
		score += 10
	}

	// Bonus for shorter names (more specific matches)
	if len(item.Name) < 20 {
		score += 5
	}

	return score
}
