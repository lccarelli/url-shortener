package main

import (
	"encoding/json"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"strconv"
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

func generateCRC32Key(url string) string {
	hash := crc32.ChecksumIEEE([]byte(url))
	return strconv.FormatUint(uint64(hash), 36) // base36 para hacerlo más corto
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	shortKey := generateCRC32Key(req.URL)

	// Intentamos guardar si no existe (SETNX)
	set, err := rdb.SetNX(ctx, shortKey, req.URL, 24*time.Hour).Result()
	if err != nil {
		http.Error(w, "error accessing Redis", http.StatusInternalServerError)
		return
	}

	if !set {
		// Ya existía, validamos que sea la misma URL
		existing, err := rdb.Get(ctx, shortKey).Result()
		if err != nil || existing != req.URL {
			http.Error(w, "hash collision detected", http.StatusConflict)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortenResponse{Short: shortKey})
}

func lookupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["short"]

	url, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

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
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		PoolSize: 1000,
	})

	router := mux.NewRouter()
	router.HandleFunc("/shorten", shortenHandler).Methods("POST")
	router.HandleFunc("/{short}", lookupHandler).Methods("GET")
	router.HandleFunc("/{short}", deleteHandler).Methods("DELETE")

	log.Println("URL Shortener service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
