package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AppInstallationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Installs a Contentful app in an environment and manages its installation parameters. The App Definition must already exist; use `contentful_app_definition` to manage a custom app definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in space_id/environment_id/app_definition_id form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space where the app is installed. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment where the app is installed. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition being installed. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"marketplace": schema.SetAttribute{
				Description: "Marketplace agreement acknowledgments sent to Contentful when installing the app. Use the acknowledgment strings required by the selected Marketplace app.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
				},
			},
			"parameters": schema.StringAttribute{
				Description: "Complete object of installation parameter values, encoded as JSON with `jsonencode(...)`. Configure every value you want to retain when updating or adopting an installation. Omitting this attribute sends no parameter object and can clear existing values on apply, including after import. Use sensitive Terraform expressions for secrets. See [Secrets and Terraform state](../guides/secrets-and-state#app-and-extension-parameters) for storage, refresh, and import limitations.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
