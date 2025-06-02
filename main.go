// @title URL Shortener API
// @version 1.0
// @description Acorta URLs
// @host localhost:8080
// @BasePath /
package main

import (
	"URLShortest/api"
	"URLShortest/config"
	"URLShortest/repository"
	"URLShortest/service"
	"context"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	config.InitLogger()
	ctx := context.Background()

	shutdown := config.InitTracer("url-shortener")
	defer shutdown(context.Background())

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		PoolSize: 1000,
	})
	
	router := mux.NewRouter()
	repo := repository.NewRedisRepository(rdb, ctx)

	shortener := service.NewShortenerService(repo, 24*time.Hour)
	handler := api.NewShortenHandler(shortener)
	router.Handle("/shorten", otelhttp.NewHandler(http.HandlerFunc(handler.Handle), "ShortenHandler")).Methods("POST")

	resolver := service.NewResolverService(repo)
	redirectHandler := api.NewRedirectHandler(resolver)
	router.Handle("/redirect/{short}", otelhttp.NewHandler(http.HandlerFunc(redirectHandler.Handle), "Redirect")).Methods("GET")

	remover := service.NewRemoverService(repo)
	deleteHandler := api.NewDeleteHandler(remover)
	router.Handle("/shorten", otelhttp.NewHandler(http.HandlerFunc(deleteHandler.Handle), "Delete")).Methods("DELETE")

	statsService := service.NewStatsService(repo)
	statsHandler := api.NewStatsHandler(statsService)
	router.Handle("/stats", otelhttp.NewHandler(http.HandlerFunc(statsHandler.Handle), "Stats")).Methods("GET")

	log.Println("Running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}
