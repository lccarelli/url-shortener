package api

import (
	"encoding/json"
	"net/http"

	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/service"

	"go.opentelemetry.io/otel"
)

// DeleteHandler Handle godoc
// @Summary Elimina una o varias URLs cortas
// @Description Permite borrar una lista de claves cortas del sistema
// @Tags delete
// @Accept json
// @Produce json
// @Param request body model.DeleteRequest true "Lista de claves a eliminar"
// @Success 200 {object} model.DeleteResponse
// @Success 207 {object} model.DeleteResponse
// @Failure 400 {string} string "invalid request"
// @Failure 500 {string} string "internal error"
// @Router /shorten [delete]
type DeleteHandler struct {
	service *service.RemoverService
}

func NewDeleteHandler(svc *service.RemoverService) *DeleteHandler {
	return &DeleteHandler{service: svc}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("delete-handler").Start(r.Context(), "HandleDelete")
	defer span.End()

	var req model.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Keys) == 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.service.DeleteMany(ctx, req.Keys)
	if err != nil {
		config.Log.WithError(err).Error("Delete failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if len(resp.NotFound) > 0 {
		status = http.StatusMultiStatus
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
