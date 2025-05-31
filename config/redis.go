package config

import (
	"github.com/go-redis/redis/v8"
)

// NewRedisClient crea una conexión Redis con una dirección dada (host:port).
func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
