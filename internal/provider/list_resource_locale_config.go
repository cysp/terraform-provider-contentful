package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type localeListResourceConfig struct {
	SpaceID       types.String            `tfsdk:"space_id"`
	EnvironmentID types.String            `tfsdk:"environment_id"`
	Order         TypedList[types.String] `tfsdk:"order"`
}

func (c localeListResourceConfig) requestParams() (cm.GetLocalesParams, diag.Diagnostics) {
	spaceID, spaceIDDiags := knownListResourceString(path.Root("space_id"), c.SpaceID)
	environmentID, environmentIDDiags := knownListResourceString(path.Root("environment_id"), c.EnvironmentID)
	order, orderDiags := knownOptionalListResourceStringList(path.Root("order"), c.Order)

	if order == nil {
		order = []string{"sys.id"}
	} else {
		filteredOrder := make([]string, 0, len(order))
		for _, value := range order {
			if value != "" {
				filteredOrder = append(filteredOrder, value)
			}
		}

		order = filteredOrder
	}

	diags := diag.Diagnostics{}
	diags.Append(spaceIDDiags...)
	diags.Append(environmentIDDiags...)
	diags.Append(orderDiags...)

	if diags.HasError() {
		return cm.GetLocalesParams{}, diags
	}

	return cm.GetLocalesParams{
		SpaceID:       spaceID,
		EnvironmentID: environmentID,
		Order:         order,
	}, diags
}

func LocaleListResourceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List Contentful Locales.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list locales.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list locales.",
				Required:    true,
			},
			"order": listschema.ListAttribute{
				Description: "Order locales by one or more attributes.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
		},
	}
}
