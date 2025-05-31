package service

import (
	"context"

	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/repository"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type RemoverService struct {
	repo *repository.RedisRepository
}

func NewRemoverService(repo *repository.RedisRepository) *RemoverService {
	return &RemoverService{repo: repo}
}

func (s *RemoverService) DeleteMany(ctx context.Context, keys []string) (*model.DeleteResponse, error) {
	ctx, span := otel.Tracer("remover-service").Start(ctx, "DeleteMany")
	defer span.End()

	var deleted, notFound []string
	traceID := span.SpanContext().TraceID().String()

	for _, key := range keys {
		found, err := s.repo.Exists(key)
		if err != nil {
			config.Log.WithError(err).WithFields(logrus.Fields{
				"key":     key,
				"traceID": traceID,
			}).Warn("Redis access error during delete")
			continue
		}

		if !found {
			config.Log.WithFields(logrus.Fields{
				"key":     key,
				"traceID": traceID,
			}).Info("Short code not found to delete")
			notFound = append(notFound, key)
			continue
		}

		err = s.repo.Delete(key)
		if err != nil {
			config.Log.WithError(err).WithFields(logrus.Fields{
				"key":     key,
				"traceID": traceID,
			}).Warn("Failed to delete short code")
			continue
		}

		config.Log.WithFields(logrus.Fields{
			"key":     key,
			"traceID": traceID,
		}).Info("Short code deleted successfully")

		deleted = append(deleted, key)
	}

	config.Log.WithFields(logrus.Fields{
		"deleted_count": len(deleted),
		"not_found":     len(notFound),
		"traceID":       traceID,
	}).Info("Finished DeleteMany execution")

	if len(deleted) > 0 {
		go func() {
			_ = s.repo.IncrementGlobalStatBy("deleted", int64(len(deleted)))
		}()
	}

	return &model.DeleteResponse{
		Deleted:  deleted,
		NotFound: notFound,
	}, nil
}
