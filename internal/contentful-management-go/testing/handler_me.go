package testing

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
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("The authenticated user could not be found."), nil), nil
	default:
		return ts.me, nil
	}
}

func (ts *Handler) SetMe(me *cm.User) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.me = me
}
