package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceTypeIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
	ResourceTypeID  types.String `tfsdk:"resource_type_id"`
}

type ResourceTypeModel struct {
	ResourceTypeIdentityModel

	ID                  types.String              `tfsdk:"id"`
	ResourceProviderID  types.String              `tfsdk:"resource_provider_id"`
	Name                types.String              `tfsdk:"name"`
	DefaultFieldMapping *ResourceTypeFieldMapping `tfsdk:"default_field_mapping"`
}

type ResourceTypeFieldMapping struct {
	Title       types.String                   `tfsdk:"title"`
	Subtitle    types.String                   `tfsdk:"subtitle"`
	Description types.String                   `tfsdk:"description"`
	ExternalURL types.String                   `tfsdk:"external_url"`
	Image       *ResourceTypeFieldMappingImage `tfsdk:"image"`
	Badge       *ResourceTypeFieldMappingBadge `tfsdk:"badge"`
}

type ResourceTypeFieldMappingImage struct {
	URL     types.String `tfsdk:"url"`
	AltText types.String `tfsdk:"alt_text"`
}

type ResourceTypeFieldMappingBadge struct {
	Label   types.String `tfsdk:"label"`
	Variant types.String `tfsdk:"variant"`
}
