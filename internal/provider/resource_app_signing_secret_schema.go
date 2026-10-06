package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const appSigningValueDescription = "Symmetric key shared between Contentful and an app backend. Must be exactly 64 characters matching `^[0-9a-zA-Z+/=_-]+$`. Exactly one of `value` and `value_wo` is required. Stored in Terraform state. Import leaves `value` null; applying the configured value can rotate the secret. Timeout-only updates preserve the secret and stored value."

type appSigningSecretValueValidator struct{}

func (appSigningSecretValueValidator) Description(_ context.Context) string {
	return "app signing secret must be exactly 64 characters and contain only letters, digits, +, /, =, _, or -"
}

func (v appSigningSecretValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (appSigningSecretValueValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(validateAppSigningSecretValue(req.ConfigValue.ValueString(), req.Path)...)
}

func validateAppSigningSecretValue(value string, valuePath path.Path) diag.Diagnostics {
	request := cm.AppSigningSecretRequestData{Value: value}
	if request.Validate() != nil {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Invalid app signing secret value", "The app signing secret must be exactly 64 characters and contain only letters, digits, +, /, =, _, or -.")}
	}

	return nil
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
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					appSigningSecretValueValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRoot("value_wo")),
				},
			},
			"value_wo": schema.StringAttribute{
				Description:         signingSecretWriteOnlyDescription,
				MarkdownDescription: signingSecretWriteOnlyDescription + " See [write-only signing secrets](../guides/secrets-and-state#write-only-signing-secrets) for rotation and saved-plan behavior.",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators:          []validator.String{appSigningSecretValueValidator{}},
			},
			"created_at": schema.StringAttribute{
				Description: "Contentful creation timestamp in RFC 3339 format, or null when omitted. It can change when the signing secret is replaced. Write-only updates can make this unknown during planning and trigger downstream replacement even without a secret write.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Contentful update timestamp in RFC 3339 format, or null when omitted. Write-only updates can make this unknown during planning and trigger downstream replacement even without a secret write.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
