package cmtesting

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type Server struct {
	server *cm.Server

	h *Handler

	rateLimiter *secondRateLimiter
}

var _ http.Handler = (*Server)(nil)

func NewContentfulManagementServer(opts ...ServerOption) (*Server, error) {
	cfg, err := buildServerConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server options: %w", err)
	}

	handler := NewHandler()
	handler.omitEntryMutationResponseFields = cfg.omitEntryMutationResponseFields

	server, err := cm.NewServer(
		handler,
		handler,
		cm.WithNotFound(notFoundHandler),
		cm.WithErrorHandler(errorHandler),
		cm.WithMiddleware(handler.authorizeAppRequest),
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var rateLimiter *secondRateLimiter
	if cfg.rateLimitPerSecond > 0 {
		rateLimiter = newSecondRateLimiter(cfg.rateLimitPerSecond, cfg.rateLimitNow)
	}

	return &Server{
		server:      server,
		h:           handler,
		rateLimiter: rateLimiter,
	}, nil
}

func (s *Server) Handler() *Handler {
	return s.h
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	limitState := s.currentRateLimitState()
	limitState.writeHeaders(w.Header())

	if !limitState.allowed {
		message := "Rate limit exceeded"

		_ = WriteContentfulManagementErrorResponse(w, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

		return
	}

	s.server.ServeHTTP(w, r)
}

func (s *Server) RegisterSpaceEnvironment(spaceID, environmentID string) {
	s.RegisterSpaceEnvironmentWithStatus(spaceID, environmentID, "ready")
}

func (s *Server) RegisterSpaceEnvironmentWithStatus(spaceID, environmentID, status string) {
	s.h.RegisterSpaceEnvironment(spaceID, environmentID, status)
}
