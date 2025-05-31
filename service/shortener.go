package service

import (
	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/repository"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"hash/crc32"
	"strconv"
	"time"
)

type ShortenerService struct {
	repo *repository.RedisRepository
	ttl  time.Duration
}

func NewShortenerService(repo *repository.RedisRepository, ttl time.Duration) *ShortenerService {
	return &ShortenerService{
		repo: repo,
		ttl:  ttl,
	}
}

func generateKey(url string) string {
	hash := crc32.ChecksumIEEE([]byte(url))
	return strconv.FormatUint(uint64(hash), 36)
}

func (s *ShortenerService) ShortenURL(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, error) {
	tr := otel.Tracer("shortener-service")
	ctx, span := tr.Start(ctx, "ShortenURL")
	defer span.End()

	if req.URL == "" {
		config.Log.WithFields(logrus.Fields{
			"url": req.URL,
		}).Warn("Received empty URL")
		return nil, ErrInvalidURL
	}

	key := generateKey(req.URL)
	config.Log.WithFields(logrus.Fields{
		"url": req.URL,
		"key": key,
	}).Info("Generated short key")

	ok, err := s.repo.SetIfNotExists(key, req.URL, s.ttl)
	if err != nil {
		config.Log.WithError(err).WithField("key", key).Error("Error writing to Redis")
		return nil, err
	}

	if !ok {
		existing, err := s.repo.Get(key)
		if err != nil || existing != req.URL {
			config.Log.WithFields(logrus.Fields{
				"key":      key,
				"existing": existing,
				"new":      req.URL,
			}).Warn("Hash collision detected")
			return nil, ErrHashCollision
		}
	}

	if ok {
		go func() {
			_ = s.repo.AddToGlobalSet(key)
			_ = s.repo.InitRanking(key)
			_ = s.repo.IncrementGlobalStat("created")
		}()
	}

	config.Log.WithFields(logrus.Fields{
		"short": key,
	}).Info("Shorten successful")

	return &model.ShortenResponse{Short: key}, nil
}

var (
	ErrInvalidURL    = errors.New("invalid URL")
	ErrHashCollision = errors.New("hash collision detected")
)
