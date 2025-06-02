package api

import (
	"net/http"

	"URLShortest/config"
	"URLShortest/service"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
)

// RedirectHandler Handle godoc
// @Summary Redirige a la URL original
// @Description Busca la URL larga a partir de un código corto y redirige con 302
// @Tags redirect
// @Produce plain
// @Param short path string true "Clave corta"
// @Success 302 {string} string "Found"
// @Failure 404 {string} string "URL no encontrada"
// @Failure 500 {string} string "Error interno"
// @Router /{short} [get]
type RedirectHandler struct {
	service *service.ResolverService
}

func NewRedirectHandler(svc *service.ResolverService) *RedirectHandler {
	return &RedirectHandler{service: svc}
}

func (h *RedirectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("redirect-handler").Start(r.Context(), "HandleRedirect")
	defer span.End()

	vars := mux.Vars(r)
	short := vars["short"]

	url, err := h.service.Resolve(ctx, short)
	if err != nil {
		if err == service.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		config.Log.WithError(err).WithField("short", short).Error("Failed to resolve short URL")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.Log.WithField("short", short).Info("Redirecting")
	println(url)
	http.Redirect(w, r, url, http.StatusFound)
}
