package service

import (
	"URLShortest/model"
	"URLShortest/repository"
	"context"

	"go.opentelemetry.io/otel"
)

// StatsService es el servicio encargado de recopilar estadísticas generales del sistema.
type StatsService struct {
	repo *repository.RedisRepository
}

// NewStatsService crea una nueva instancia del servicio de estadísticas.
func NewStatsService(repo *repository.RedisRepository) *StatsService {
	return &StatsService{repo: repo}
}

// GetGeneralStats obtiene un resumen general de estadísticas del sistema.
// Incluye total de URLs acortadas, cantidad total de accesos, estadísticas por código,
// accesos recientes y los códigos más visitados.
func (s *StatsService) GetGeneralStats(ctx context.Context) (*model.GeneralStatsResponse, error) {
	// Iniciar traza con OpenTelemetry.
	ctx, span := otel.Tracer("stats-service").Start(ctx, "GetGeneralStats")
	defer span.End()

	// Obtener todos los códigos acortados del sistema.
	allKeys, err := s.repo.GetAllShortCodes()
	if err != nil {
		return nil, err
	}

	var totalVisits int
	var recent []model.ShortStatsEntry

	// Recorrer cada short code y acumular estadísticas individuales.
	for _, short := range allKeys {
		visits, err := s.repo.GetVisitCount(short)
		if err != nil {
			// Si no se puede obtener, se ignora esa entrada.
			continue
		}
		totalVisits += visits

		lastAccess, _ := s.repo.GetLastAccess(short)
		recent = append(recent, model.ShortStatsEntry{
			ShortCode:  short,
			Visits:     visits,
			LastAccess: lastAccess,
		})
	}

	// Obtener los 5 códigos más accedidos.
	top, _ := s.repo.GetTopAccessed(5)

	// Obtener métricas globales de creación, eliminación y resolución.
	created, _ := s.repo.GetGlobalStat("created")
	deleted, _ := s.repo.GetGlobalStat("deleted")
	resolved, _ := s.repo.GetGlobalStat("resolved")

	// Devolver la respuesta con todas las métricas agregadas.
	return &model.GeneralStatsResponse{
		TotalShortened: len(allKeys),
		TotalVisits:    totalVisits,
		Created:        created,
		Deleted:        deleted,
		Resolved:       resolved,
		TopAccessed:    top,
		RecentAccesses: recent,
	}, nil
}
