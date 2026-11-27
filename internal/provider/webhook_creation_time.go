package provider

import (
	"context"
	"encoding/json"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const createdAtPrivateKey = "created_at"

func readWebhookCreationTime(ctx context.Context, private PrivateProviderData) (*time.Time, diag.Diagnostics) {
	data, diags := private.GetKey(ctx, createdAtPrivateKey)
	if len(data) == 0 || diags.HasError() {
		return nil, diags
	}

	var prior *time.Time

	err := json.Unmarshal(data, &prior)
	if err != nil || prior == nil {
		diags.AddError("Invalid private webhook creation time", "Restore valid Terraform state or re-import the webhook to observe its creation time again.")
	}

	return prior, diags
}

// A timestamp is evidence of recreation only when both instants are known.
// A version decrease breaks continuity independently, even without a timestamp.
type webhookCreationObservation struct {
	recreated bool
	changed   bool
	createdAt *time.Time
}

func observeWebhookCreation(prior *time.Time, version types.Int64, sys cm.WebhookDefinitionSys) webhookCreationObservation {
	lower := !version.IsNull() && !version.IsUnknown() && int64(sys.Version) < version.ValueInt64()

	observed, present := sys.CreatedAt.Get()
	if !present {
		return webhookCreationObservation{recreated: lower, changed: lower && prior != nil}
	}

	changed := prior == nil || !prior.Equal(observed)
	utc := observed.UTC()

	return webhookCreationObservation{recreated: lower || (prior != nil && changed), changed: changed, createdAt: &utc}
}

func (o webhookCreationObservation) publish(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	if !o.changed {
		return nil
	}

	if o.createdAt == nil {
		return private.SetKey(ctx, createdAtPrivateKey, nil)
	}

	return SetPrivateProviderData(ctx, private, createdAtPrivateKey, o.createdAt)
}

func webhookWithoutCredentialHistory(model WebhookModel) WebhookModel {
	model.HTTPBasicUsername = types.StringNull()

	model.HTTPBasicPassword = types.StringNull()
	if !model.Headers.IsNull() {
		model.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{})
	}

	return model
}

// Project the new lifetime without inheriting ordinary secrets from the plan.
// Where a reply omits descriptors, guard only write-only values sent to the new lifetime.
func recreatedWebhookUpdateState(ctx context.Context, sanitized cm.WebhookDefinition, prior, plan, submitted WebhookModel, passwordDiscarded bool) (WebhookModel, diag.Diagnostics) {
	data, diags := NewWebhookResourceModelFromResponse(ctx, sanitized, webhookWithoutCredentialHistory(plan))
	if data.HTTPBasicUsername.IsNull() {
		sentWriteOnly := plan.HTTPBasicPassword.IsNull() && !submitted.HTTPBasicPassword.IsNull()
		switch {
		case sentWriteOnly && (passwordDiscarded || (sanitized.HttpBasicUsername.IsEmpty() && sanitized.HttpBasicPassword.IsEmpty())):
			data.HTTPBasicUsername = submitted.HTTPBasicUsername
		case passwordDiscarded:
			data.HTTPBasicUsername = prior.HTTPBasicUsername
			if data.HTTPBasicUsername.IsNull() {
				data.HTTPBasicUsername = plan.HTTPBasicUsername
			}
		}
	}

	if sanitized.Headers == nil {
		headers := make(map[string]TypedObject[WebhookHeaderValue])

		for key, sent := range submitted.Headers.Elements() {
			if !sent.Value().Value.IsNull() && plan.Headers.Elements()[key].Value().Value.IsNull() {
				headers[key] = NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true)})
			}
		}

		if len(headers) != 0 {
			data.Headers = NewTypedMap(headers)
		}
	}

	data.Timeouts = plan.Timeouts

	return data, diags
}
