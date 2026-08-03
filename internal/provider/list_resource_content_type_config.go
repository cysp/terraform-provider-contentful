package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type contentTypeListResourceConfig struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

func (c contentTypeListResourceConfig) requestParams() (cm.GetContentTypesParams, diag.Diagnostics) {
	spaceID, spaceIDDiags := knownListResourceString(path.Root("space_id"), c.SpaceID)
	environmentID, environmentIDDiags := knownListResourceString(path.Root("environment_id"), c.EnvironmentID)

	diags := diag.Diagnostics{}
	diags.Append(spaceIDDiags...)
	diags.Append(environmentIDDiags...)

	if diags.HasError() {
		return cm.GetContentTypesParams{}, diags
	}

	return cm.GetContentTypesParams{
		SpaceID:       spaceID,
		EnvironmentID: environmentID,
		Order:         []string{"sys.id"},
	}, diags
}

func ContentTypeListResourceConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List Contentful Content Types.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list content types.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list content types.",
				Required:    true,
			},
		},
	}
}
