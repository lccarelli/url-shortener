package integration

import (
	"URLShortest/api"
	"URLShortest/model"
	"URLShortest/repository"
	"URLShortest/service"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteHandler(t *testing.T) {
	repo := repository.NewDefaultRedisRepository("localhost:6379")
	shortener := service.NewShortenerService(repo, 10*time.Minute)
	remover := service.NewRemoverService(repo)
	handler := api.NewDeleteHandler(remover)

	ctx := context.Background()

	// Preparamos una URL válida para delete_existing
	respDeletable, err := shortener.ShortenURL(ctx, model.ShortenRequest{URL: "https://deletable.com"})
	assert.NoError(t, err)
	validKey := respDeletable.Short

	// Preparamos una URL válida para delete_mixed
	respMixed, err := shortener.ShortenURL(ctx, model.ShortenRequest{URL: "https://mixed.com"})
	assert.NoError(t, err)
	mixedKey := respMixed.Short

	invalidKey := "non_existent_123"

	tests := []struct {
		name           string
		keys           []string
		expectedStatus int
		expectedDelete []string
		expectedMiss   []string
	}{
		{
			name:           "delete_existing",
			keys:           []string{validKey},
			expectedStatus: http.StatusOK,
			expectedDelete: []string{validKey},
		},
		{
			name:           "delete_nonexistent",
			keys:           []string{invalidKey},
			expectedStatus: http.StatusMultiStatus,
			expectedMiss:   []string{invalidKey},
		},
		{
			name:           "delete_mixed",
			keys:           []string{mixedKey, invalidKey},
			expectedStatus: http.StatusMultiStatus,
			expectedDelete: []string{mixedKey},
			expectedMiss:   []string{invalidKey},
		},
		{
			name:           "invalid_empty_request",
			keys:           []string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := model.DeleteRequest{Keys: tc.keys}
			jsonData, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodDelete, "/shorten", bytes.NewBuffer(jsonData))
			rec := httptest.NewRecorder()

			handler.Handle(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if rec.Code == http.StatusOK || rec.Code == http.StatusMultiStatus {
				var response model.DeleteResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.ElementsMatch(t, tc.expectedDelete, response.Deleted)
				assert.ElementsMatch(t, tc.expectedMiss, response.NotFound)
			}
		})
	}
}
