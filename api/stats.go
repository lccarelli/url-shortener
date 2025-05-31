package api

import (
	"encoding/json"
	"net/http"

	"URLShortest/config"
	"URLShortest/service"

	"go.opentelemetry.io/otel"
)

// StatsHandler Handle godoc
// @Summary Obtiene estadísticas generales del sistema
// @Description Devuelve el total de URLs acortadas, accesos totales y más accedidas
// @Tags stats
// @Produce json
// @Success 200 {object} model.GeneralStatsResponse
// @Failure 500 {string} string "internal error"
// @Router /stats [get]
type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{service: svc}
}

func (h *StatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("stats-handler").Start(r.Context(), "HandleStats")
	defer span.End()

	resp, err := h.service.GetGeneralStats(ctx)
	if err != nil {
		config.Log.WithError(err).Error("Failed to get stats")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
