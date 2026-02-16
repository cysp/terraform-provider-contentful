package cmtesting

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAuthenticatedUser(_ context.Context) (cm.GetAuthenticatedUserRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	switch ts.me {
	case nil:
		return NewContentfulManagementErrorStatusCodeNotFound(new("The authenticated user could not be found."), nil), nil
	default:
		return ts.me, nil
	}
}
