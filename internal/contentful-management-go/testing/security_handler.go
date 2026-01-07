package testing

import (
	"context"
	"errors"
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type SecurityHandler struct {
	mu sync.Mutex
}

const ValidAccessToken = "CFPAT-12345"

var ErrAccessTokenInvalid = errors.New("AccessTokenInvalid")

var _ cm.SecurityHandler = (*SecurityHandler)(nil)

func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{
		mu: sync.Mutex{},
	}
}

func (h *SecurityHandler) HandleAccessToken(ctx context.Context, _ cm.OperationName, token cm.AccessToken) (context.Context, error) {
	if token.Token != ValidAccessToken {
		return nil, ErrAccessTokenInvalid
	}

	return ctx, nil
}
