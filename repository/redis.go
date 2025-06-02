package repository

import (
	"URLShortest/model"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepository(client *redis.Client, ctx context.Context) *RedisRepository {
	return &RedisRepository{client: client, ctx: ctx}
}

func NewDefaultRedisRepository(addr string) *RedisRepository {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisRepository{client: client, ctx: context.Background()}
}

func (r *RedisRepository) SetIfNotExists(key string, value string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, ttl).Result()
}

func (r *RedisRepository) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisRepository) IncrementVisitStats(short string) error {
	pipe := r.client.TxPipeline()

	pipe.Incr(r.ctx, "stats:"+short+":count")
	pipe.Set(r.ctx, "stats:"+short+":last_access", time.Now().Format(time.RFC3339), 0)

	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisRepository) Exists(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

func (r *RedisRepository) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisRepository) GetAllShortCodes() ([]string, error) {
	return r.client.SMembers(r.ctx, "short:all").Result()
}

func (r *RedisRepository) GetVisitCount(short string) (int, error) {
	return r.client.Get(r.ctx, "stats:"+short+":count").Int()
}

func (r *RedisRepository) GetLastAccess(short string) (string, error) {
	return r.client.Get(r.ctx, "stats:"+short+":last_access").Result()
}

func (r *RedisRepository) GetTopAccessed(n int) ([]model.ShortStatsEntry, error) {
	entries, err := r.client.ZRevRangeWithScores(r.ctx, "stats:ranking", 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	var result []model.ShortStatsEntry
	for _, entry := range entries {
		result = append(result, model.ShortStatsEntry{
			ShortCode: entry.Member.(string),
			Visits:    int(entry.Score),
		})
	}
	return result, nil
}

func (r *RedisRepository) AddToGlobalSet(short string) error {
	return r.client.SAdd(r.ctx, "short:all", short).Err()
}

func (r *RedisRepository) InitRanking(short string) error {
	return r.client.ZAdd(r.ctx, "stats:ranking", &redis.Z{
		Score:  0,
		Member: short,
	}).Err()
}

func (r *RedisRepository) IncrementRanking(short string) error {
	return r.client.ZIncrBy(r.ctx, "stats:ranking", 1, short).Err()
}

func (r *RedisRepository) IncrementGlobalStat(key string) error {
	return r.client.Incr(r.ctx, "stats:global:"+key).Err()
}

func (r *RedisRepository) IncrementGlobalStatBy(key string, amount int64) error {
	return r.client.IncrBy(r.ctx, "stats:global:"+key, amount).Err()
}

func (r *RedisRepository) GetGlobalStat(key string) (int, error) {
	return r.client.Get(r.ctx, "stats:global:"+key).Int()
}

func (r *RedisRepository) Set(key string, value string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}
