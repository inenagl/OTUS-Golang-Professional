package internalhttp

import (
	"net/http"
	"time"

	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

func (s Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		nrw := negroni.NewResponseWriter(w)

		next.ServeHTTP(nrw, r)

		latency := time.Since(t)
		s.logger.Info("Request processed",
			zap.String("IP", r.RemoteAddr),
			zap.Time("datetime", t),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("HTTP-version", r.Proto),
			zap.Int("response code", nrw.Status()),
			zap.Duration("latency", latency),
			zap.String("user-agent", r.Header.Get("User-Agent")),
		)
	})
}
