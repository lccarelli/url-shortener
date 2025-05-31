package api

import (
	"encoding/json"
	"net/http"

	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/service"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
)

// LookupHandler Handle godoc
// @Summary Consulta una URL corta
// @Description Devuelve la URL original asociada a una clave corta
// @Tags lookup
// @Produce json
// @Param short path string true "Clave corta"
// @Success 200 {object} model.LookupResponse
// @Failure 404 {string} string "URL no encontrada"
// @Failure 500 {string} string "Error interno"
// @Router /lookup/{short} [get]
type LookupHandler struct {
	service *service.ResolverService
}

func NewLookupHandler(svc *service.ResolverService) *LookupHandler {
	return &LookupHandler{service: svc}
}

func (h *LookupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("lookup-handler").Start(r.Context(), "HandleLookup")
	defer span.End()

	vars := mux.Vars(r)
	short := vars["short"]

	url, err := h.service.Resolve(ctx, short)
	if err != nil {
		if err == service.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		config.Log.WithError(err).WithField("short", short).Error("Failed to lookup short URL")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := model.LookupResponse{URL: url}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
