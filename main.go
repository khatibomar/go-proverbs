package main

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-proverbs/go-proverbs/internal/proverbs"
	"github.com/go-proverbs/go-proverbs/internal/web"
)

//go:embed internal/proverbs/examples/official/*.gotmpl internal/proverbs/examples/community/*.gotmpl
var exampleFS embed.FS

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Set the embedded filesystem for examples
	proverbs.SetExampleFS(exampleFS)

	// Load proverbs
	collection := proverbs.LoadAllProverbs()
	logger.Info("loaded proverbs", "total", len(collection.GetAll()))

	// Validate collection
	if errors := collection.ValidateCollection(); len(errors) > 0 {
		logger.Warn("validation errors found", "count", len(errors))
		for _, err := range errors {
			logger.Warn("validation error", "error", err.Error())
		}
	}

	// Create web handler
	webHandler := web.NewHandler(collection, logger)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API routes
	mux.HandleFunc("GET /api/v1/proverbs", handleGetProverbs(collection))
	mux.HandleFunc("GET /api/v1/proverbs/random", handleGetRandomProverb(collection))
	mux.HandleFunc("GET /api/v1/proverbs/search", handleSearchProverbs(collection))
	mux.HandleFunc("GET /api/v1/proverbs/stats", handleGetStats(collection))
	mux.HandleFunc("GET /api/v1/proverbs/categories/{category}", handleGetByCategory(collection))
	mux.HandleFunc("GET /api/v1/proverbs/sources/{source}", handleGetBySource(collection))
	mux.HandleFunc("GET /api/v1/proverbs/tags/{tag}", handleGetByTag(collection))

	// Web UI routes
	mux.HandleFunc("GET /proverbs/{id}", webHandler.HandleProverb)
	mux.HandleFunc("GET /categories", webHandler.HandleCategories)
	mux.HandleFunc("GET /categories/{category}", webHandler.HandleCategory)
	mux.HandleFunc("GET /tags", webHandler.HandleTags)
	mux.HandleFunc("GET /tags/{tag}", webHandler.HandleTag)
	mux.HandleFunc("GET /sources/{source}", webHandler.HandleSource)
	mux.HandleFunc("GET /search", webHandler.HandleSearch)
	mux.HandleFunc("GET /random", webHandler.HandleRandom)
	mux.HandleFunc("GET /", webHandler.HandleIndex)

	// Apply middleware
	handler := loggingMiddleware(logger)(corsMiddleware(mux))

	// Server configuration
	port := getEnvOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}

// API Handlers

func handleGetProverbs(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := getIntParam(r, "limit", 50)
		offset := getIntParam(r, "offset", 0)

		allProverbs := collection.GetAll()
		total := len(allProverbs)

		// Apply pagination
		start := offset
		end := offset + limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		response := map[string]any{
			"proverbs": allProverbs[start:end],
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		}

		writeJSONResponse(w, response)
	}
}

func handleGetRandomProverb(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proverb := collection.GetRandomProverb()
		writeJSONResponse(w, proverb)
	}
}

func handleSearchProverbs(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		results := collection.SearchProverbs(query)
		response := map[string]any{
			"query":   query,
			"results": results,
			"count":   len(results),
		}

		writeJSONResponse(w, response)
	}
}

func handleGetStats(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := collection.GetStats()
		writeJSONResponse(w, stats)
	}
}

func handleGetByCategory(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryStr := r.PathValue("category")
		category := proverbs.Category(categoryStr)

		results := collection.GetByCategory(category)
		response := map[string]any{
			"category": category,
			"proverbs": results,
			"count":    len(results),
		}

		writeJSONResponse(w, response)
	}
}

func handleGetBySource(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceStr := r.PathValue("source")
		source := proverbs.Source(sourceStr)

		results := collection.GetBySource(source)
		response := map[string]any{
			"source":   source,
			"proverbs": results,
			"count":    len(results),
		}

		writeJSONResponse(w, response)
	}
}

func handleGetByTag(collection *proverbs.ProverbCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.PathValue("tag")

		results := collection.GetByTag(tag)
		response := map[string]any{
			"tag":      tag,
			"proverbs": results,
			"count":    len(results),
		}

		writeJSONResponse(w, response)
	}
}

// Middleware

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", duration,
			)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Utility functions

func writeJSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func getIntParam(r *http.Request, param string, defaultValue int) int {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}