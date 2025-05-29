// main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Short string `json:"short"`
}

// generateShortKeyNoRedis codifica un ID en base62
func generateShortKey(id int64) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	short := ""
	for id > 0 {
		short = string(chars[id%62]) + short
		id /= 62
	}
	return short
}

// shortenHandler maneja POST /shorten
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	id, err := rdb.Incr(ctx, "url_id").Result()
	if err != nil {
		http.Error(w, "error generating ID", http.StatusInternalServerError)
		return
	}

	shortKey := generateShortKey(id)
	err = rdb.Set(ctx, shortKey, req.URL, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "error saving URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortenResponse{Short: shortKey})
}

// resolveHandler maneja GET /{short}
func resolveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	short := vars["short"]

	url, err := rdb.Get(ctx, short).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// deleteHandler maneja DELETE /{short}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	short := vars["short"]

	err := rdb.Del(ctx, short).Err()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// --- NUEVO: tunear transporte global ---
	t := http.DefaultTransport.(*http.Transport)
	t.MaxIdleConns = 10000
	t.MaxIdleConnsPerHost = 10000
	t.IdleConnTimeout = 90 * time.Second

	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	router := mux.NewRouter()
	router.HandleFunc("/shorten", shortenHandler).Methods("POST")
	router.HandleFunc("/{short}", resolveHandler).Methods("GET")
	router.HandleFunc("/{short}", deleteHandler).Methods("DELETE")

	log.Println("URL Shortener service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
