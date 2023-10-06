package internalhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

type ContextKey string

const UserIDKey ContextKey = "currentUserId"

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

func (s Server) userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-API-User")
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uid, err := uuid.Parse(userID)
		if err != nil {
			s.logger.Error(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			s.writeError(w, "userId is not valid UUID")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, uid)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
