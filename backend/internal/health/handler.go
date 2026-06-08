package health

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	dbStatus := "up"
	status := http.StatusOK
	if err := h.pool.Ping(context.Background()); err != nil {
		dbStatus = "down"
		status = http.StatusServiceUnavailable
	}
	httpx.WriteJSON(w, status, map[string]string{"status": "ok", "db": dbStatus})
}

func (h *Handler) Metrics() http.Handler {
	return promhttp.Handler()
}
