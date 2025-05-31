package service

import (
	"URLShortest/config"
	"URLShortest/repository"
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var ErrNotFound = errors.New("short URL not found")

type ResolverService struct {
	repo *repository.RedisRepository
}

func NewResolverService(repo *repository.RedisRepository) *ResolverService {
	return &ResolverService{repo: repo}
}

func (s *ResolverService) Resolve(ctx context.Context, short string) (string, error) {
	tr := otel.Tracer("resolver-service")
	ctx, span := tr.Start(ctx, "Resolve")
	defer span.End()

	url, err := s.repo.Get(short)
	if err != nil {
		if err.Error() == "redis: nil" {
			config.Log.WithFields(logrus.Fields{
				"short": short,
				"trace": span.SpanContext().TraceID().String(),
			}).Warn("Short code not found in Redis")
			return "", ErrNotFound
		}
		config.Log.WithError(err).WithFields(logrus.Fields{
			"short": short,
			"trace": span.SpanContext().TraceID().String(),
		}).Error("Failed to fetch short code from Redis")
		return "", err
	}

	config.Log.WithFields(logrus.Fields{
		"short": short,
		"url":   url,
		"trace": span.SpanContext().TraceID().String(),
	}).Info("Resolved short URL successfully")

	// Asynchronous update of stats
	go func() {
		_ = s.repo.IncrementVisitStats(short)
		_ = s.repo.IncrementRanking(short)
		_ = s.repo.IncrementGlobalStat("resolved")
		
		err1 := s.repo.IncrementVisitStats(short)
		if err1 != nil {
			config.Log.WithError(err1).WithField("short", short).Warn("Failed to increment visit stats")
		}

		err2 := s.repo.IncrementRanking(short)
		if err2 != nil {
			config.Log.WithError(err2).WithField("short", short).Warn("Failed to increment ranking")
		}
	}()

	return url, nil
}
