package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const appSigningValueDescription = "Symmetric key shared between Contentful and an app backend. Must be exactly 64 characters matching `^[0-9a-zA-Z+/=_-]+$`. Stored in Terraform state. Import leaves `value` null; applying the configured value can rotate the secret. Timeout-only updates preserve the secret and stored value."

const appSigningSecretValueConstraint = "app signing secret must be exactly 64 characters and contain only letters, digits, +, /, =, _, or -"

type appSigningSecretValueValidator struct{}

func (appSigningSecretValueValidator) Description(_ context.Context) string {
	return appSigningSecretValueConstraint
}

func (v appSigningSecretValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (appSigningSecretValueValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	_, diags := newAppSigningSecretRequest(req.ConfigValue.ValueString(), req.Path)
	resp.Diagnostics.Append(diags...)
}

func AppSigningSecretResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Signing Secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in `organization_id/app_definition_id` form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the app. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition for which the signing secret is created. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description:         appSigningValueDescription,
				MarkdownDescription: appSigningValueDescription + " See [importing signing secrets](../guides/secrets-and-state#importing-signing-secrets).",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					appSigningSecretValueValidator{},
				},
			},
			"timeouts": timeouts.Attributes(ctx, crudTimeoutOptions()),
		},
	}
}
