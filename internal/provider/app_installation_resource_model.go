package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppInstallationResourceModel struct {
	ID              types.String         `tfsdk:"id"`
	SpaceID         types.String         `tfsdk:"space_id"`
	EnvironmentID   types.String         `tfsdk:"environment_id"`
	AppDefinitionID types.String         `tfsdk:"app_definition_id"`
	Marketplace     types.Set            `tfsdk:"marketplace"`
	Parameters      jsontypes.Normalized `tfsdk:"parameters"`
}

func AppInstallationResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"marketplace": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"parameters": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
			},
		},
	}
}
