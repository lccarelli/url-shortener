package service

import (
	"context"

	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/repository"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// RemoverService es el servicio encargado de eliminar claves cortas de Redis.
type RemoverService struct {
	repo *repository.RedisRepository
}

// NewRemoverService crea una nueva instancia del servicio de eliminación.
func NewRemoverService(repo *repository.RedisRepository) *RemoverService {
	return &RemoverService{repo: repo}
}

// DeleteMany elimina múltiples claves cortas del almacenamiento Redis.
// Retorna una lista de claves eliminadas exitosamente y otra de claves no encontradas.
func (s *RemoverService) DeleteMany(ctx context.Context, keys []string) (*model.DeleteResponse, error) {
	// Iniciar traza con OpenTelemetry.
	ctx, span := otel.Tracer("remover-service").Start(ctx, "DeleteMany")
	defer span.End()

	var deleted, notFound []string
	traceID := span.SpanContext().TraceID().String()

	// Recorrer cada clave y realizar el proceso de eliminación.
	for _, key := range keys {
		// Verificar si la clave existe en Redis.
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

		// Eliminar la clave de Redis.
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

	// Log final con resumen de la operación.
	config.Log.WithFields(logrus.Fields{
		"deleted_count": len(deleted),
		"not_found":     len(notFound),
		"traceID":       traceID,
	}).Info("Finished DeleteMany execution")

	// Actualizar métricas globales de eliminaciones (de forma asíncrona).
	if len(deleted) > 0 {
		go func() {
			_ = s.repo.IncrementGlobalStatBy("deleted", int64(len(deleted)))
		}()
	}

	// Devolver respuesta con resultados de la operación.
	return &model.DeleteResponse{
		Deleted:  deleted,
		NotFound: notFound,
	}, nil
}
