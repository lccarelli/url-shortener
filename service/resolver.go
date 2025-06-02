package service

import (
	"URLShortest/config"
	"URLShortest/repository"
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// ErrNotFound se devuelve cuando no se encuentra la URL acortada en Redis.
var ErrNotFound = errors.New("short URL not found")

// ResolverService es el servicio encargado de resolver códigos cortos a sus URLs originales.
type ResolverService struct {
	repo *repository.RedisRepository
}

// NewResolverService crea una nueva instancia del servicio de resolución.
func NewResolverService(repo *repository.RedisRepository) *ResolverService {
	return &ResolverService{repo: repo}
}

// Resolve busca una URL original asociada a un código corto.
// Si se encuentra, actualiza las estadísticas de acceso en segundo plano.
func (s *ResolverService) Resolve(ctx context.Context, short string) (string, error) {
	tr := otel.Tracer("resolver-service")
	ctx, span := tr.Start(ctx, "Resolve")
	defer span.End()

	// Buscar la URL original en Redis.
	url, err := s.repo.Get(short)
	if err != nil {
		if err.Error() == "redis: nil" {
			// La clave no existe en Redis.
			config.Log.WithFields(logrus.Fields{
				"short": short,
				"trace": span.SpanContext().TraceID().String(),
			}).Warn("Short code not found in Redis")
			return "", ErrNotFound
		}

		// Otro tipo de error al acceder a Redis.
		config.Log.WithError(err).WithFields(logrus.Fields{
			"short": short,
			"trace": span.SpanContext().TraceID().String(),
		}).Error("Failed to fetch short code from Redis")
		return "", err
	}

	// Log de resolución exitosa.
	config.Log.WithFields(logrus.Fields{
		"short": short,
		"url":   url,
		"trace": span.SpanContext().TraceID().String(),
	}).Info("Resolved short URL successfully")

	// Actualización asíncrona de estadísticas de acceso.
	go func() {
		err1 := s.repo.IncrementVisitStats(short)
		if err1 != nil {
			config.Log.WithError(err1).WithField("short", short).Warn("Failed to increment visit stats")
		}

		err2 := s.repo.IncrementRanking(short)
		if err2 != nil {
			config.Log.WithError(err2).WithField("short", short).Warn("Failed to increment ranking")
		}

		err3 := s.repo.IncrementGlobalStat("resolved")
		if err3 != nil {
			config.Log.WithError(err3).WithField("short", short).Warn("Failed to increment global resolved stat")
		}
	}()

	return url, nil
}
