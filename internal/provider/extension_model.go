package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExtensionIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ExtensionID   types.String `tfsdk:"extension_id"`
}

type ExtensionModel struct {
	ExtensionIdentityModel

	ID         types.String             `tfsdk:"id"`
	Extension  *ExtensionModelExtension `tfsdk:"extension"`
	Parameters jsontypes.Normalized     `tfsdk:"parameters"`
}

type ExtensionModelExtension struct {
	Name       types.String                          `tfsdk:"name"`
	FieldTypes []AppDefinitionLocationFieldTypesItem `tfsdk:"field_types"`
	Src        types.String                          `tfsdk:"src"`
	SrcDoc     types.String                          `tfsdk:"srcdoc"`
	Parameters *AppDefinitionParameters              `tfsdk:"parameters"`
	Sidebar    types.Bool                            `tfsdk:"sidebar"`
}

type ExtensionModelExtensionFieldType struct {
	Type     types.String                           `tfsdk:"type"`
	LinkType types.String                           `tfsdk:"link_type"`
	Items    *ExtensionModelExtensionFieldTypeItems `tfsdk:"items"`
}

type ExtensionModelExtensionFieldTypeItems struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}
