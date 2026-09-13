package cmtesting

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetWebhookSigningSecret(_ context.Context, params cm.GetWebhookSigningSecretParams) (cm.GetWebhookSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// The master environment is the mock's existing space-presence fixture.
	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	secret, ok := ts.webhookSigningSecrets[params.SpaceID]
	if !ok {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"WebhookSigningSecret does not exist."`)), nil
	}

	return &secret, nil
}

//nolint:ireturn
func (ts *Handler) PutWebhookSigningSecret(_ context.Context, req *cm.WebhookSigningSecretRequestData, params cm.PutWebhookSigningSecretParams) (cm.PutWebhookSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	if req.Validate() != nil {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil), nil
	}

	_, exists := ts.webhookSigningSecrets[params.SpaceID]

	secret := cm.WebhookSigningSecret{
		Sys: cm.WebhookSigningSecretSys{
			Type:  cm.WebhookSigningSecretSysTypeWebhookSigningSecret,
			Space: cm.NewSpaceLink(params.SpaceID),
		},
		RedactedValue: req.Value[len(req.Value)-4:],
	}
	ts.webhookSigningSecrets[params.SpaceID] = secret

	if exists {
		return (*cm.PutWebhookSigningSecretOK)(&secret), nil
	}

	return (*cm.PutWebhookSigningSecretCreated)(&secret), nil
}

//nolint:ireturn
func (ts *Handler) DeleteWebhookSigningSecret(_ context.Context, params cm.DeleteWebhookSigningSecretParams) (cm.DeleteWebhookSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	if _, ok := ts.webhookSigningSecrets[params.SpaceID]; !ok {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, []byte(`"WebhookSigningSecret does not exist."`)), nil
	}

	delete(ts.webhookSigningSecrets, params.SpaceID)

	return &cm.NoContent{}, nil
}
