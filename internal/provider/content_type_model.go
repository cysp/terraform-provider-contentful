package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContentTypeModel struct {
	ID            types.String                                  `tfsdk:"id"`
	SpaceID       types.String                                  `tfsdk:"space_id"`
	EnvironmentID types.String                                  `tfsdk:"environment_id"`
	ContentTypeID types.String                                  `tfsdk:"content_type_id"`
	Name          types.String                                  `tfsdk:"name"`
	Description   types.String                                  `tfsdk:"description"`
	DisplayField  types.String                                  `tfsdk:"display_field"`
	Fields        TypedList[TypedObject[ContentTypeFieldValue]] `tfsdk:"fields"`
	Metadata      ContentTypeMetadataValue                      `tfsdk:"metadata"`
}

type ContentTypeFieldValue struct {
	ID               types.String                                                     `tfsdk:"id"`
	Name             types.String                                                     `tfsdk:"name"`
	FieldType        types.String                                                     `tfsdk:"type"`
	LinkType         types.String                                                     `tfsdk:"link_type"`
	Disabled         types.Bool                                                       `tfsdk:"disabled"`
	Omitted          types.Bool                                                       `tfsdk:"omitted"`
	Required         types.Bool                                                       `tfsdk:"required"`
	DefaultValue     jsontypes.Normalized                                             `tfsdk:"default_value"`
	Items            TypedObject[ContentTypeFieldItemsValue]                          `tfsdk:"items"`
	Localized        types.Bool                                                       `tfsdk:"localized"`
	Validations      TypedList[jsontypes.Normalized]                                  `tfsdk:"validations"`
	AllowedResources TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]] `tfsdk:"allowed_resources"`
}

type ContentTypeFieldItemsValue struct {
	ItemsType   types.String                    `tfsdk:"type"`
	LinkType    types.String                    `tfsdk:"link_type"`
	Validations TypedList[jsontypes.Normalized] `tfsdk:"validations"`
}

type ContentTypeFieldAllowedResourceItemValue struct {
	ContentfulEntry TypedObject[ContentTypeFieldAllowedResourceItemContentfulEntryValue] `tfsdk:"contentful_entry"`
	External        TypedObject[ContentTypeFieldAllowedResourceItemExternalValue]        `tfsdk:"external"`
}

type ContentTypeFieldAllowedResourceItemContentfulEntryValue struct {
	Source       types.String            `tfsdk:"source"`
	ContentTypes TypedList[types.String] `tfsdk:"content_types"`
}

type ContentTypeFieldAllowedResourceItemExternalValue struct {
	TypeID types.String `tfsdk:"type"`
}
