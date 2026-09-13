package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookSigningSecretResourceModelFromResponse(response cm.WebhookSigningSecret, spaceID string, value types.String) (WebhookSigningSecretModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// The endpoint identifies the singleton. Never redirect state to an
	// unexpected space or interpret redacted metadata as secret equality.
	if response.Sys.Space.Sys.ID == "" || response.Sys.Space.Sys.ID != spaceID {
		diags.AddAttributeError(path.Root("space_id"), "Unexpected webhook signing secret identity", "Contentful did not return the requested space identity. No new state was published.")
	}

	if value.IsUnknown() {
		diags.AddAttributeError(path.Root("value"), "Unexpected unknown signing secret value", "The retained signing secret value must be known or null.")
	}

	if diags.HasError() {
		return WebhookSigningSecretModel{}, diags
	}

	return WebhookSigningSecretModel{
		IDIdentityModel:                   NewIDIdentityModelFromMultipartID(response.Sys.Space.Sys.ID),
		WebhookSigningSecretIdentityModel: WebhookSigningSecretIdentityModel{SpaceID: types.StringValue(response.Sys.Space.Sys.ID)},
		Value:                             value,
	}, diags
}
