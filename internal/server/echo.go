package server

import (
	"fmt"
	"net/http"
	"strings"

	libmessage "github.com/forwardhttp/go-lib/message"
	"github.com/pkg/errors"
)

func (s *server) handleEcho(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleEcho")

	var ctx = r.Context()

	switch r.Method {
	case http.MethodConnect, http.MethodTrace:
		s.writeError(ctx, w, http.StatusMethodNotAllowed, nil)
		return
	}

	message := &libmessage.ConsumerMessage{
		Headers: r.Header,
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

	s.writeResponse(ctx, w, http.StatusOK, message)

}
