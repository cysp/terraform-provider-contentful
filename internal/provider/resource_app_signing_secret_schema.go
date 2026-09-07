package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const appSigningValueDescription = "The symmetric key shared between Contentful and an app backend. Must be exactly 64 characters and match `^[0-9a-zA-Z+/=_-]+$`. The complete value is stored in Terraform state when the provider writes a secret to Contentful. Timeout-only updates preserve the remote secret and the stored value."

var appSigningSecretValuePattern = regexp.MustCompile(`^[0-9a-zA-Z+/=_-]+$`)

type appSigningSecretValueValidator struct{}

func (appSigningSecretValueValidator) Description(_ context.Context) string {
	return "app signing secret must be exactly 64 characters and contain only letters, digits, +, /, =, _, or -"
}

func (v appSigningSecretValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v appSigningSecretValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if len(value) == 64 && appSigningSecretValuePattern.MatchString(value) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid app signing secret value",
		v.Description(ctx)+".",
	)
}

func AppSigningSecretResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Signing Secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in organization_id/app_definition_id form.",
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
				Description:         appSigningValueDescription + " Import cannot recover the existing secret and leaves value null. Without ignore_changes = [value], applying the configured value replaces the remote secret; a configuration-driven import can do this during the import apply. With ignore_changes = [value], the imported value remains null, including after timeout changes.",
				MarkdownDescription: appSigningValueDescription + " Import cannot recover the existing secret and leaves `value` null. Without `ignore_changes = [value]`, applying the configured value replaces the remote secret; a configuration-driven import can do this during the import apply. With `ignore_changes = [value]`, the imported value remains null, including after timeout changes. See [Secrets and Terraform state](../guides/secrets-and-state) for storage, refresh, and import guidance.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					appSigningSecretValueValidator{},
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
