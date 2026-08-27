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

const appSigningValueDescription = "The symmetric key shared between Contentful and an app backend. Must be exactly 64 characters and match `^[0-9a-zA-Z+/=_-]+$`. `Sensitive` masks routine Terraform CLI and HCP Terraform UI output. After Create or Update, Terraform stores the complete configured value in resource state, and saved plan files can also contain it. Protect access to both."

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
		Description: "Manages a Contentful App Signing Secret. Contentful returns only `redactedValue` for the secret, not the complete submitted value. Refresh preserves the value already in Terraform state and cannot detect an out-of-band replacement. An imported resource initially has a null value because the complete current value cannot be recovered; the next configured apply writes and stores the configured replacement. Protect state and saved plan artifacts; state protection depends on the configured Terraform backend.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the app.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition for which the signing secret is created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: appSigningValueDescription,
				Required:    true,
				Sensitive:   true,
				Validators: []validator.String{
					appSigningSecretValueValidator{},
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
