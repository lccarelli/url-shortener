package api

import (
	"URLShortest/model"
	"URLShortest/service"
	"encoding/json"
	"net/http"
)

// ShortenHandler es el handler HTTP para acortar URLs largas.
type ShortenHandler struct {
	service *service.ShortenerService
}

// NewShortenHandler crea una nueva instancia del handler.
func NewShortenHandler(svc *service.ShortenerService) *ShortenHandler {
	return &ShortenHandler{service: svc}
}

// Handle godoc
// @Summary Acorta una URL larga
// @Description Genera una clave corta para redireccionar hacia la URL original
// @Tags shortener
// @Accept json
// @Produce json
// @Param request body model.ShortenRequest true "URL larga a acortar"
// @Success 200 {object} model.ShortenResponse
// @Failure 400 {string} string "invalid request"
// @Failure 409 {string} string "hash collision detected"
// @Failure 500 {string} string "internal error"
// @Router /shorten [post]
func (h *ShortenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest

	// valida request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// manda al servicio
	// TODO: No debe generarse la misma url corta dada una larga, usar un timestamp o algo similar
	// TODO Agregar que las keys expiren despues de x tiempo 10min?
	resp, err := h.service.ShortenURL(r.Context(), req)
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrHashCollision:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	// devuelve respuesta
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
