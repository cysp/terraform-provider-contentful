package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const webhookSigningValueDescription = "Symmetric key shared between Contentful and webhook receivers in the space. Must be exactly 64 characters matching `^[0-9a-zA-Z+/=_-]+$`. Stored in Terraform state. Refresh cannot detect external rotation. Import leaves `value` null; applying the configured value can rotate the secret. Timeout-only updates preserve the secret and stored value."

const webhookSigningSecretValueConstraint = "The webhook signing secret must be exactly 64 characters matching ^[0-9a-zA-Z+/=_-]+$."

type webhookSigningSecretValueValidator struct{}

func (webhookSigningSecretValueValidator) Description(_ context.Context) string {
	return webhookSigningSecretValueConstraint
}

func (v webhookSigningSecretValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (webhookSigningSecretValueValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(validateWebhookSigningSecretValue(req.ConfigValue.ValueString(), req.Path)...)
}

func validateWebhookSigningSecretValue(value string, valuePath path.Path) diag.Diagnostics {
	request := cm.WebhookSigningSecretRequestData{Value: value}
	diags := diag.Diagnostics{}

	// OpenAPI owns the format constraint. Never expose the validator's raw error.
	if request.Validate() != nil {
		diags.AddAttributeError(valuePath, "Invalid webhook signing secret value", webhookSigningSecretValueConstraint)
	}

	return diags
}

func WebhookSigningSecretResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages the Contentful Webhook Signing Secret for a space. Creating this resource replaces any existing secret. Destroying it deletes the current secret, including one rotated outside Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform resource identifier, equal to `space_id`.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"space_id": schema.StringAttribute{
				Description:   "ID of the space for which the signing secret is created. Changing this value replaces the resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"value": schema.StringAttribute{
				Description:         webhookSigningValueDescription,
				MarkdownDescription: webhookSigningValueDescription + " See [importing signing secrets](../guides/secrets-and-state#importing-signing-secrets).",
				Required:            true,
				Sensitive:           true,
				Validators:          []validator.String{webhookSigningSecretValueValidator{}},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
