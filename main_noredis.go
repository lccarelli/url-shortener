package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type shortenRequestNoRedis struct {
	URL string `json:"url"`
}
type shortenResponseNoRedis struct {
	Short string `json:"short"`
}

var store = make(map[string]string)
var mu sync.RWMutex

func generateShortKeyNoRedis(id int64) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := ""
	for id > 0 {
		key = string(chars[id%62]) + key
		id /= 62
	}
	return key
}

func main() {
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		log.Println("hello")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[1:]
		mu.RLock()
		url, ok := store[key]
		mu.RUnlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	})

	log.Println("Listening on :8080")
	srv := &http.Server{
		Addr:              ":8080",
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Fatal(srv.ListenAndServe())
}
