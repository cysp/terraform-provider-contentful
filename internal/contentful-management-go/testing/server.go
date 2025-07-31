package testing

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type Server struct {
	server *cm.Server

	h   *Handler
	sec *SecurityHandler
}

var _ http.Handler = (*Server)(nil)

func NewContentfulManagementServer() (*Server, error) {
	handler := NewHandler()

	securityHandler := NewSecurityHandler()

	server, err := cm.NewServer(handler, securityHandler, cm.WithNotFound(notFoundHandler))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &Server{
		server: server,
		h:      handler,
		sec:    securityHandler,
	}, nil
}

func (s *Server) Handler() *Handler {
	return s.h
}

func (s *Server) SecurityHandler() *SecurityHandler {
	return s.sec
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}
