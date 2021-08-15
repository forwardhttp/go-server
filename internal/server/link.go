package server

import (
	"fmt"
	"net/http"
	"strings"

	libmessage "github.com/forwardhttp/go-lib/message"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func (s *server) handlePostNew(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	link, err := s.connection.CreateConnection(ctx)
	if err != nil {
		s.writeError(ctx, w, 0, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, link)

}

func (s *server) handleOpenLink(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	hash := chi.URLParam(r, "hash")
	if hash == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("hash is required but not defined"))
		return
	}

	err := s.connection.Connect(ctx, hash, w, r)
	if err != nil {
		s.writeError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Debug("connection has been opened")

}

func (s *server) handleLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleLink")
	var ctx = r.Context()

	hash, route := chi.URLParam(r, "hash"), chi.URLParam(r, "*")
	if hash == "" {
		s.writeError(ctx, w, http.StatusNotFound, errors.New("hash is required but not defined"))
		return
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	switch r.Method {
	case http.MethodConnect, http.MethodTrace:
		s.writeError(ctx, w, http.StatusMethodNotAllowed, nil)
		return
	}

	message := &libmessage.ConsumerMessage{
		Headers: r.Header,
		Route:   route,
		Method:  r.Method,
	}

	if bodyAllowedForMethod(r.Method) {
		defer s.closeRequestBody(ctx, r)
		err := message.ReadFromRequest(r)
		if err != nil {
			if strings.Contains(err.Error(), "http: request body too large") {
				s.writeError(ctx, w, http.StatusRequestEntityTooLarge, errors.New("requests cannot be larger than 4096 bytes"))
				return
			}

			s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to read request body"))
			return
		}
	}

	err := s.connection.PushMessage(ctx, hash, message)
	if err != nil {
		s.writeError(ctx, w, http.StatusInternalServerError, errors.Wrap(err, "failed to forward payload"))
		return
	}

	s.writeResponse(ctx, w, http.StatusAccepted, nil)

}

func bodyAllowedForMethod(method string) bool {
	for _, m := range []string{http.MethodPatch, http.MethodPut, http.MethodPost, http.MethodDelete} {
		if m == method {
			return true
		}
	}

	return false

}
