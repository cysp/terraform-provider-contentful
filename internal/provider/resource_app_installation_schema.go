package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AppInstallationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version:     1,
		Description: "Installs a Contentful app in an environment and manages its installation parameters.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in `space_id/environment_id/app_definition_id` form.",
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
				Description: "Marketplace agreement acknowledgments sent to Contentful when installing the app. Defaults to an empty set. Use the acknowledgment strings required by the selected Marketplace app.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(emptyAppInstallationMarketplace()),
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
				},
			},
			"parameters": schema.MapAttribute{
				Description: "Installation parameters keyed by parameter ID. Use `jsonencode` for each value. See [Updating parameters](#updating-parameters) for the effects of omitting or ignoring a key.",
				ElementType: jsontypes.NormalizedType{},
				Validators:  []validator.Map{installationParameterValuesValidator{}},
				Optional:    true,
			},
			"timeouts": timeouts.Attributes(ctx, crudTimeoutOptions()),
		},
	}
}
