package testing

import (
	"context"
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type SecurityHandler struct {
	mu sync.Mutex
}

var _ cm.SecurityHandler = (*SecurityHandler)(nil)

func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{
		mu: sync.Mutex{},
	}
}

func (h *SecurityHandler) HandleAccessToken(ctx context.Context, _ cm.OperationName, _ cm.AccessToken) (context.Context, error) {
	return ctx, nil
}
