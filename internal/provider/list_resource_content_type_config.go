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
	spaceID, spaceIDDiags := requestRequiredString(c.SpaceID, path.Root("space_id"))
	environmentID, environmentIDDiags := requestRequiredString(c.EnvironmentID, path.Root("environment_id"))

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
		Description: "Lists Contentful Content Types in an existing space and environment.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "ID of the space from which to list Content Types.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "ID of the environment from which to list Content Types.",
				Required:    true,
			},
		},
	}
}
