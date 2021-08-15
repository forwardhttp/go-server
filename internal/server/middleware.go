package server

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// NewStructuredLogger is a constructor for creating a request logger middleware
func (s *server) requestLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&structuredLogger{logger})
}

func (s *server) maxByteReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)

		next.ServeHTTP(w, r)
	})
}
