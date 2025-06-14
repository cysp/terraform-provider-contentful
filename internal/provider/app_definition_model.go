package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppDefinitionModel struct {
	ID              types.String                 `tfsdk:"id"`
	OrganizationID  types.String                 `tfsdk:"organization_id"`
	AppDefinitionID types.String                 `tfsdk:"app_definition_id"`
	Name            types.String                 `tfsdk:"name"`
	Src             types.String                 `tfsdk:"src"`
	BundleID        types.String                 `tfsdk:"bundle_id"`
	Locations       []AppDefinitionLocationsItem `tfsdk:"locations"`
	Parameters      *AppDefinitionParameters     `tfsdk:"parameters"`
}

type AppDefinitionLocationsItem struct {
	Location       types.String                          `tfsdk:"location"`
	FieldTypes     []AppDefinitionLocationFieldTypesItem `tfsdk:"field_types"`
	NavigationItem *AppDefinitionLocationNavigationItem  `tfsdk:"navigation_item"`
}

type AppDefinitionLocationFieldTypesItem struct {
	Type     types.String                             `tfsdk:"type"`
	LinkType types.String                             `tfsdk:"link_type"`
	Items    *AppDefinitionLocationFieldTypeItemsItem `tfsdk:"items"`
}

type AppDefinitionLocationFieldTypeItemsItem struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}

type AppDefinitionLocationNavigationItem struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

type AppDefinitionParameters struct {
	Installation []AppDefinitionParameter `tfsdk:"installation"`
	Instance     []AppDefinitionParameter `tfsdk:"instance"`
}

type AppDefinitionParameter struct {
	ID          string                          `tfsdk:"id"`
	Type        string                          `tfsdk:"type"`
	Name        string                          `tfsdk:"name"`
	Description *string                         `tfsdk:"description"`
	Required    *bool                           `tfsdk:"required"`
	Default     jsontypes.Normalized            `tfsdk:"default"`
	Options     TypedList[jsontypes.Normalized] `tfsdk:"options"`
	Labels      *AppDefinitionParameterLabels   `tfsdk:"labels"`
}

type AppDefinitionParameterLabels struct {
	Empty *string `tfsdk:"empty"`
	True  *string `tfsdk:"true"`
	False *string `tfsdk:"false"`
}
