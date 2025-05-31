package service

import (
	"URLShortest/model"
	"URLShortest/repository"
	"context"

	"go.opentelemetry.io/otel"
)

type StatsService struct {
	repo *repository.RedisRepository
}

func NewStatsService(repo *repository.RedisRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetGeneralStats(ctx context.Context) (*model.GeneralStatsResponse, error) {
	ctx, span := otel.Tracer("stats-service").Start(ctx, "GetGeneralStats")
	defer span.End()

	allKeys, err := s.repo.GetAllShortCodes()
	if err != nil {
		return nil, err
	}

	var totalVisits int
	var recent []model.ShortStatsEntry

	for _, short := range allKeys {
		visits, err := s.repo.GetVisitCount(short)
		if err != nil {
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

	top, _ := s.repo.GetTopAccessed(5)

	// NUEVO: métricas globales
	created, _ := s.repo.GetGlobalStat("created")
	deleted, _ := s.repo.GetGlobalStat("deleted")
	resolved, _ := s.repo.GetGlobalStat("resolved")

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
