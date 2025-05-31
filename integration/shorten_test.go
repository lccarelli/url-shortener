package integration

import (
	"URLShortest/api"
	"URLShortest/model"
	"URLShortest/repository"
	"URLShortest/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var handler *api.ShortenHandler

func TestMain(m *testing.M) {
	repo := repository.NewDefaultRedisRepository("localhost:6379")
	service := service.NewShortenerService(repo, 10*time.Minute)
	handler = api.NewShortenHandler(service)

	code := m.Run()
	os.Exit(code)
}

func TestShorten_Success(t *testing.T) {
	body := model.ShortenRequest{URL: "https://example.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	handler.Handle(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.ShortenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Short)
}

func TestShorten_Parametrized(t *testing.T) {
	testCases := map[string]string{
		"basic_example":    "https://example.com",
		"long_url":         "https://www.google.com/search?q=openai+gpt&sourceid=chrome",
		"trailing_slash":   "https://example.com/",
		"subdomain":        "https://blog.example.com/article",
		"with_query_param": "https://example.com/page?utm_source=chatgpt",
	}

	for name, url := range testCases {
		t.Run(name, func(t *testing.T) {
			body := model.ShortenRequest{URL: url}
			jsonBody, err := json.Marshal(body)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
			rec := httptest.NewRecorder()

			handler.Handle(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var response model.ShortenResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Short)
		})
	}
}
