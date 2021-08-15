package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pe "github.com/forwardhttp/go-lib/errors"
	"github.com/forwardhttp/go-server/internal/connection"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type server struct {
	port       uint
	connection connection.Service
	server     *http.Server
	logger     *logrus.Logger
}

func New(port uint, logger *logrus.Logger, connection connection.Service) *server {
	r := chi.NewRouter()

	s := &server{
		port:       port,
		connection: connection,
		logger:     logger,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadTimeout:       time.Second * 5,
			ReadHeaderTimeout: time.Second * 5,
			WriteTimeout:      time.Second * 5,
			MaxHeaderBytes:    512,
		},
	}

	s.buildRouter(r)
	return s
}

func (s *server) buildRouter(r *chi.Mux) {
	r.Use(s.requestLogger(s.logger))
	r.Use(middleware.SetHeader("content-type", "application/json"))
	r.Post("/new", s.handlePostNew)
	r.Get("/open/{hash}", s.handleOpenLink)
	r.Group(func(r chi.Router) {
		r.Use(s.maxByteReader)
		r.Handle("/echo", http.HandlerFunc(s.handleEcho))
		r.Handle("/{hash}", http.HandlerFunc(s.handleLink))
		r.Handle("/{hash}/*", http.HandlerFunc(s.handleLink))
	})
}

func (s *server) Run() error {
	s.logger.WithField("service", "Server").Infof("Starting on Port %d", s.port)
	return s.server.ListenAndServe()
}

// GracefullyShutdown gracefully shuts down the HTTP API.
func (s *server) GracefullyShutdown(ctx context.Context) error {
	s.logger.Info("attempting to shutdown server gracefully")
	return s.server.Shutdown(ctx)
}

func (s *server) closeRequestBody(ctx context.Context, r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		s.logger.WithError(err).Error("failed to close request body")
	}
}

func (s *server) writeResponse(ctx context.Context, w http.ResponseWriter, code int, data interface{}) {

	if code != http.StatusOK {
		w.WriteHeader(code)
	}

	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (s *server) writeError(ctx context.Context, w http.ResponseWriter, code int, err error) {

	// If err is not nil, actually pass in a map so that the output to the wire is {"error": "text...."} else just let it fall through
	if err != nil {

		payload := map[string]interface{}{
			"message": err.Error(),
		}
		var peError pe.Error
		if errors.As(err, &peError) {
			switch peError.T {
			case pe.InternalErrorType:
				s.writeResponse(ctx, w, http.StatusInternalServerError, nil)
			case pe.ValidateErrorType:
				s.writeResponse(ctx, w, http.StatusBadRequest, payload)
			}
			return
		}

		s.writeResponse(ctx, w, code, payload)
		return

	}

	s.writeResponse(ctx, w, code, nil)

}
