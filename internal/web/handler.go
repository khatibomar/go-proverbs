package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-proverbs/go-proverbs/internal/proverbs"
)

// Handler handles web requests
type Handler struct {
	collection *proverbs.ProverbCollection
	logger     *slog.Logger
	templates  *template.Template
}

// NewHandler creates a new web handler
func NewHandler(collection *proverbs.ProverbCollection, logger *slog.Logger) *Handler {
	templates := template.Must(template.New("").Funcs(templateFuncs).ParseGlob("web/templates/*.html"))

	return &Handler{
		collection: collection,
		logger:     logger,
		templates:  templates,
	}
}

// HandleIndex serves the main index page
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	stats := h.collection.GetStats()
	
	// Get a mix of official and community proverbs for the homepage
	var recentProverbs []proverbs.Proverb
	official := h.collection.GetBySource(proverbs.SourceOfficial)
	community := h.collection.GetBySource(proverbs.SourceCommunity)
	
	// Take first 5 official and first 5 community
	for i := 0; i < 5 && i < len(official); i++ {
		recentProverbs = append(recentProverbs, official[i])
	}
	for i := 0; i < 5 && i < len(community); i++ {
		recentProverbs = append(recentProverbs, community[i])
	}

	data := PageData{
		Title:        "Go Proverbs: Official & Community Edition",
		Description:  "A comprehensive collection of Go programming wisdom",
		TemplateName: "index-content",
		Stats:        stats,
		Proverbs:     h.toProverbsWithID(recentProverbs),
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "index.html", data)
}

// HandleProverb serves a single proverb page
func (h *Handler) HandleProverb(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	foundProverb := h.collection.GetByID(id)
	if foundProverb == nil {
		http.NotFound(w, r)
		return
	}

	// Create ProverbWithID for the found proverb
	proverbWithID := &ProverbWithID{
		Proverb: *foundProverb,
		ID:      id,
	}

	// Get related proverbs (same category)
	related := h.collection.GetByCategory(foundProverb.Category)
	var relatedFiltered []proverbs.Proverb
	for _, p := range related {
		// Skip the current proverb by comparing all fields since we don't have ID
		if p.Title != foundProverb.Title && len(relatedFiltered) < 5 {
			relatedFiltered = append(relatedFiltered, p)
		}
	}

	// Get all proverbs for navigation
	allProverbs := h.collection.GetAll()
	var foundIndex int = -1
	for i, proverb := range allProverbs {
		if proverb.Title == foundProverb.Title && proverb.Text == foundProverb.Text {
			foundIndex = i
			break
		}
	}

	// Get previous and next proverbs
	var prevProverb, nextProverb *ProverbWithID
	if foundIndex > 0 {
		prev := h.toProverbWithID(allProverbs[foundIndex-1])
		prevProverb = &prev
	}
	if foundIndex >= 0 && foundIndex < len(allProverbs)-1 {
		next := h.toProverbWithID(allProverbs[foundIndex+1])
		nextProverb = &next
	}

	data := PageData{
		Title:        foundProverb.Title,
		Description:  foundProverb.Text,
		TemplateName: "proverb-content",
		Proverb:      proverbWithID,
		Proverbs:     h.toProverbsWithID(relatedFiltered),
		PrevProverb:  prevProverb,
		NextProverb:  nextProverb,
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "proverb.html", data)
}

// HandleCategories serves the categories overview page
func (h *Handler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	stats := h.collection.GetStats()

	h.logger.Info("handling categories request", "path", r.URL.Path, "stats_categories", len(stats.Categories))
	for category, count := range stats.Categories {
		h.logger.Info("category stats", "category", category, "count", count)
	}

	data := PageData{
		Title:        "Categories - Go Proverbs",
		Description:  "Browse Go proverbs by category",
		TemplateName: "categories-content",
		Stats:        stats,
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "categories.html", data)
}

// HandleCategory serves proverbs for a specific category
func (h *Handler) HandleCategory(w http.ResponseWriter, r *http.Request) {
	categoryStr := r.PathValue("category")
	category := proverbs.Category(categoryStr)

	h.logger.Info("handling category request", "category_str", categoryStr, "category", category, "path", r.URL.Path)

	categoryProverbs := h.collection.GetByCategory(category)
	h.logger.Info("category proverbs found", "category", category, "count", len(categoryProverbs))

	// Debug: log all available categories
	allProverbs := h.collection.GetAll()
	categoryMap := make(map[string]int)
	for _, p := range allProverbs {
		categoryMap[string(p.Category)]++
	}
	h.logger.Info("all available categories", "categories", categoryMap)

	data := PageData{
		Title:        fmt.Sprintf("%s - Go Proverbs", strings.Title(string(category))),
		Description:  fmt.Sprintf("Go proverbs about %s", category),
		TemplateName: "category-content",
		Category:     string(category),
		Proverbs:     h.toProverbsWithID(categoryProverbs),
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "category.html", data)
}

// HandleTags displays all available tags
func (h *Handler) HandleTags(w http.ResponseWriter, r *http.Request) {
	stats := h.collection.GetStats()
	
	data := PageData{
		Title:        "All Tags - Go Proverbs",
		Description:  "Browse Go proverbs by tags",
		TemplateName: "tags-content",
		Stats:        stats,
		CurrentYear:  time.Now().Year(),
	}
	
	h.renderTemplate(w, "tags.html", data)
}

// HandleTag displays proverbs for a specific tag
func (h *Handler) HandleTag(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	
	proverbList := h.collection.GetByTag(tag)
	
	data := PageData{
		Title:        fmt.Sprintf("Tag: %s - Go Proverbs", tag),
		Description:  fmt.Sprintf("Go proverbs tagged with %s", tag),
		TemplateName: "tag-content",
		Tag:          tag,
		Proverbs:     h.toProverbsWithID(proverbList),
		Stats:        h.collection.GetStats(),
		CurrentYear:  time.Now().Year(),
	}
	
	h.renderTemplate(w, "tag.html", data)
}

// HandleSource serves proverbs for a specific source
func (h *Handler) HandleSource(w http.ResponseWriter, r *http.Request) {
	sourceStr := r.PathValue("source")
	source := proverbs.Source(sourceStr)

	sourceProverbs := h.collection.GetBySource(source)

	data := PageData{
		Title:        fmt.Sprintf("%s Proverbs - Go Proverbs", strings.Title(string(source))),
		Description:  fmt.Sprintf("%s Go proverbs", strings.Title(string(source))),
		TemplateName: "source-content",
		Source:       string(source),
		Proverbs:     h.toProverbsWithID(sourceProverbs),
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "source.html", data)
}

// HandleSearch serves search results
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var results []proverbs.Proverb

	if query != "" {
		results = h.collection.SearchProverbs(query)
	}

	data := PageData{
		Title:        fmt.Sprintf("Search: %s - Go Proverbs", query),
		Description:  fmt.Sprintf("Search results for '%s'", query),
		TemplateName: "search-content",
		Query:        query,
		Proverbs:     h.toProverbsWithID(results),
		CurrentYear:  time.Now().Year(),
	}

	h.renderTemplate(w, "search.html", data)
}

// HandleRandom serves a random proverb
func (h *Handler) HandleRandom(w http.ResponseWriter, r *http.Request) {
	randomProverb := h.collection.GetRandomProverb()
	// Find the ID by searching through the maps
	var proverbID string
	for id, proverb := range h.collection.Official {
		if proverb.Title == randomProverb.Title && proverb.Text == randomProverb.Text {
			proverbID = id
			break
		}
	}
	if proverbID == "" {
		for id, proverb := range h.collection.Community {
			if proverb.Title == randomProverb.Title && proverb.Text == randomProverb.Text {
				proverbID = id
				break
			}
		}
	}
	http.Redirect(w, r, "/proverbs/"+proverbID, http.StatusFound)
}

// ProverbWithID wraps a proverb with its ID for template use
type ProverbWithID struct {
	proverbs.Proverb
	ID string `json:"id"`
}

// Helper function to find proverb ID
func (h *Handler) findProverbID(proverb proverbs.Proverb) string {
	for id, p := range h.collection.Official {
		if p.Title == proverb.Title && p.Text == proverb.Text {
			return id
		}
	}
	for id, p := range h.collection.Community {
		if p.Title == proverb.Title && p.Text == proverb.Text {
			return id
		}
	}
	return ""
}

// Helper function to convert proverb to ProverbWithID
func (h *Handler) toProverbWithID(proverb proverbs.Proverb) ProverbWithID {
	return ProverbWithID{
		Proverb: proverb,
		ID:      h.findProverbID(proverb),
	}
}

// Helper function to convert slice of proverbs to ProverbWithID
func (h *Handler) toProverbsWithID(proverbs []proverbs.Proverb) []ProverbWithID {
	result := make([]ProverbWithID, len(proverbs))
	for i, proverb := range proverbs {
		result[i] = h.toProverbWithID(proverb)
	}
	return result
}

// PageData represents data passed to templates
type PageData struct {
	Title        string
	Description  string
	Query        string
	Category     string
	Source       string
	Tag          string
	TemplateName string
	Stats        proverbs.ProverbStats
	Proverb      *ProverbWithID
	Proverbs     []ProverbWithID
	PrevProverb  *ProverbWithID
	NextProverb  *ProverbWithID
	CurrentYear  int
}

// renderTemplate renders a template with the given data
func (h *Handler) renderTemplate(w http.ResponseWriter, tmpl string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	h.logger.Info("rendering template", "template", tmpl, "title", data.Title)

	// Log key data for debugging
	h.logger.Info("template data",
		"template", tmpl,
		"stats_total", data.Stats.Total,
		"stats_official", data.Stats.Official,
		"stats_community", data.Stats.Community,
		"stats_categories_count", len(data.Stats.Categories),
		"proverbs_count", len(data.Proverbs))

	if err := h.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		h.logger.Error("template execution failed", "template", tmpl, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("template rendered successfully", "template", tmpl)
}

// Template functions
var templateFuncs = template.FuncMap{
	"title": strings.Title,
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"formatCategory": func(category interface{}) string {
		var categoryStr string
		switch v := category.(type) {
		case proverbs.Category:
			categoryStr = string(v)
		case string:
			categoryStr = v
		default:
			categoryStr = fmt.Sprintf("%v", v)
		}
		return strings.ReplaceAll(strings.Title(categoryStr), "-", " ")
	},
	"formatSource": func(source interface{}) string {
		var sourceStr string
		switch v := source.(type) {
		case proverbs.Source:
			sourceStr = string(v)
		case string:
			sourceStr = v
		default:
			sourceStr = fmt.Sprintf("%v", v)
		}
		return strings.Title(sourceStr)
	},
	"formatDate": func(t time.Time) string {
		return t.Format("January 2, 2006")
	},
	"truncate": func(s string, length int) string {
		if len(s) <= length {
			return s
		}
		return s[:length] + "..."
	},
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"formatCode": func(code string) template.HTML {
		// Simple code formatting - replace newlines with <br> and add syntax highlighting classes
		formatted := strings.ReplaceAll(template.HTMLEscapeString(code), "\n", "<br>")
		formatted = strings.ReplaceAll(formatted, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;")
		return template.HTML("<pre><code class='language-go'>" + formatted + "</code></pre>")
	},
	"add": func(a, b int) int {
		return a + b
	},
}
