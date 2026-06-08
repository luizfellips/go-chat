package middleware

import (
	"net/http"
	"strconv"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/luizf/go-chat/backend/internal/metrics"
	"github.com/rs/zerolog/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		status := strconv.Itoa(ww.Status())
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration.Seconds())
		log.Info().
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Dur("duration", duration).
			Msg("request")
	})
}
