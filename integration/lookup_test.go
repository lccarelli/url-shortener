package integration

import (
	"URLShortest/api"
	"URLShortest/model"
	"URLShortest/repository"
	"URLShortest/service"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLookupHandler(t *testing.T) {
	// Setup
	repo := repository.NewDefaultRedisRepository("localhost:6379")
	shortener := service.NewShortenerService(repo, 10*time.Minute)
	resolver := service.NewResolverService(repo)
	handler := api.NewLookupHandler(resolver)

	// URLs a registrar
	testCases := map[string]string{
		"example":     "https://example.com",
		"google":      "https://google.com/search?q=shortener",
		"with_params": "https://example.com/path?utm=123",
	}

	// Crear las short URLs
	shortMap := make(map[string]string)
	for name, url := range testCases {
		resp, err := shortener.ShortenURL(context.Background(), model.ShortenRequest{URL: url})
		assert.NoError(t, err)
		shortMap[name] = resp.Short
	}

	for name, short := range shortMap {
		t.Run("resolve_"+name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/lookup/"+short, nil)
			rec := httptest.NewRecorder()

			// Necesitamos simular mux.Vars
			router := mux.NewRouter()
			router.HandleFunc("/lookup/{short}", handler.Handle)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.LookupResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, testCases[name], res.URL)
		})
	}

	t.Run("resolve_not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/lookup/not_exist_123", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/lookup/{short}", handler.Handle)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
