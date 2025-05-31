package integration

import (
	"URLShortest/api"
	"URLShortest/model"
	"URLShortest/repository"
	"URLShortest/service"
	"context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedirectHandler(t *testing.T) {
	// Setup
	repo := repository.NewDefaultRedisRepository("localhost:6379")
	shortener := service.NewShortenerService(repo, 10*time.Minute)
	resolver := service.NewResolverService(repo)
	handler := api.NewRedirectHandler(resolver)

	// Crear short URL
	originalURL := "https://redirect-test.com"
	resp, err := shortener.ShortenURL(context.Background(), model.ShortenRequest{URL: originalURL})
	assert.NoError(t, err)
	shortCode := resp.Short

	t.Run("redirect_success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/{short}", handler.Handle)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, originalURL, rec.Header().Get("Location"))
	})

	t.Run("redirect_not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent123", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/{short}", handler.Handle)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
