package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppDefinitionDataSourceModel struct {
	IDIdentityModel
	AppDefinitionIdentityModel

	Name       types.String                           `tfsdk:"name"`
	Src        types.String                           `tfsdk:"src"`
	BundleID   types.String                           `tfsdk:"bundle_id"`
	Locations  []AppDefinitionDataSourceLocationsItem `tfsdk:"locations"`
	Parameters *AppDefinitionParameters               `tfsdk:"parameters"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type AppDefinitionDataSourceLocationsItem struct {
	Location       types.String                                    `tfsdk:"location"`
	FieldTypes     []AppDefinitionDataSourceLocationFieldTypesItem `tfsdk:"field_types"`
	NavigationItem *AppDefinitionLocationNavigationItem            `tfsdk:"navigation_item"`
}

type AppDefinitionDataSourceLocationFieldTypesItem struct {
	Type     types.String                              `tfsdk:"type"`
	LinkType types.String                              `tfsdk:"link_type"`
	Items    []AppDefinitionLocationFieldTypeItemsItem `tfsdk:"items"`
}

func NewAppDefinitionDataSourceModelFromResponse(ctx context.Context, response cm.AppDefinition) (AppDefinitionDataSourceModel, diag.Diagnostics) {
	base, diags := NewAppDefinitionBaseModelFromResponse(ctx, response)

	model := AppDefinitionDataSourceModel{
		IDIdentityModel:            base.IDIdentityModel,
		AppDefinitionIdentityModel: base.AppDefinitionIdentityModel,
		Name:                       base.Name,
		Src:                        base.Src,
		BundleID:                   base.BundleID,
		Parameters:                 base.Parameters,
	}

	if base.Locations != nil {
		model.Locations = make([]AppDefinitionDataSourceLocationsItem, len(base.Locations))

		for i, location := range base.Locations {
			model.Locations[i] = AppDefinitionDataSourceLocationsItem{
				Location:       location.Location,
				NavigationItem: location.NavigationItem,
			}

			if location.FieldTypes != nil {
				model.Locations[i].FieldTypes = make([]AppDefinitionDataSourceLocationFieldTypesItem, len(location.FieldTypes))

				for fieldTypeIndex, fieldType := range location.FieldTypes {
					model.Locations[i].FieldTypes[fieldTypeIndex] = AppDefinitionDataSourceLocationFieldTypesItem{
						Type:     fieldType.Type,
						LinkType: fieldType.LinkType,
					}

					if fieldType.Items != nil {
						model.Locations[i].FieldTypes[fieldTypeIndex].Items = []AppDefinitionLocationFieldTypeItemsItem{*fieldType.Items}
					}
				}
			}
		}
	}

	return model, diags
}
