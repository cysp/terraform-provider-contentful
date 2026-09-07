package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Version zero is the state representation used by v0.0.62. Its decoding
// schema and models are independent of the current Extension and App Definition
// schemas. Validation, defaults, and plan modifiers do not run during decoding.
func extensionStateSchemaV0(ctx context.Context) schema.Schema {
	parameter := schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Required: true},
		"type":        schema.StringAttribute{Required: true},
		"name":        schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Optional: true},
		"required":    schema.BoolAttribute{Optional: true},
		"default":     schema.StringAttribute{Optional: true, CustomType: jsontypes.NormalizedType{}},
		"options": schema.ListAttribute{
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
		},
		"labels": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{Optional: true},
				"true":  schema.StringAttribute{Optional: true},
				"false": schema.StringAttribute{Optional: true},
			},
		},
	}}

	return schema.Schema{Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"space_id":       schema.StringAttribute{Required: true},
		"environment_id": schema.StringAttribute{Required: true},
		"extension_id":   schema.StringAttribute{Required: true},
		"parameters":     schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}},
		"timeouts": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"create": schema.StringAttribute{Optional: true},
				"read":   schema.StringAttribute{Optional: true},
				"update": schema.StringAttribute{Optional: true},
				"delete": schema.StringAttribute{Optional: true},
			},
		},
		"extension": schema.SingleNestedAttribute{
			Required: true,
			Attributes: map[string]schema.Attribute{
				"name":    schema.StringAttribute{Required: true},
				"src":     schema.StringAttribute{Optional: true, Computed: true},
				"srcdoc":  schema.StringAttribute{Optional: true, Computed: true},
				"sidebar": schema.BoolAttribute{Optional: true, Computed: true},
				"field_types": schema.ListNestedAttribute{
					Required: true,
					NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
						"type":      schema.StringAttribute{Required: true},
						"link_type": schema.StringAttribute{Optional: true},
						"items": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"type":      schema.StringAttribute{Required: true},
								"link_type": schema.StringAttribute{Optional: true},
							},
						},
					}},
				},
				"parameters": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"installation": schema.ListNestedAttribute{Optional: true, NestedObject: parameter},
						"instance":     schema.ListNestedAttribute{Optional: true, NestedObject: parameter},
					},
				},
			},
		},
	}}
}

type extensionStateV0 struct {
	ID            types.String              `tfsdk:"id"`
	SpaceID       types.String              `tfsdk:"space_id"`
	EnvironmentID types.String              `tfsdk:"environment_id"`
	ExtensionID   types.String              `tfsdk:"extension_id"`
	Extension     *extensionConfigurationV0 `tfsdk:"extension"`
	Parameters    jsontypes.Normalized      `tfsdk:"parameters"`
	Timeouts      types.Object              `tfsdk:"timeouts"`
}

type extensionConfigurationV0 struct {
	Name       types.String           `tfsdk:"name"`
	Src        types.String           `tfsdk:"src"`
	SrcDoc     types.String           `tfsdk:"srcdoc"`
	Sidebar    types.Bool             `tfsdk:"sidebar"`
	FieldTypes []extensionFieldTypeV0 `tfsdk:"field_types"`
	Parameters *extensionParametersV0 `tfsdk:"parameters"`
}

type extensionFieldTypeV0 struct {
	Type     types.String               `tfsdk:"type"`
	LinkType types.String               `tfsdk:"link_type"`
	Items    *extensionFieldTypeItemsV0 `tfsdk:"items"`
}

type extensionFieldTypeItemsV0 struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}

type extensionParametersV0 struct {
	Installation []extensionParameterV0 `tfsdk:"installation"`
	Instance     []extensionParameterV0 `tfsdk:"instance"`
}

type extensionParameterV0 struct {
	ID          string                          `tfsdk:"id"`
	Type        string                          `tfsdk:"type"`
	Name        string                          `tfsdk:"name"`
	Description *string                         `tfsdk:"description"`
	Required    *bool                           `tfsdk:"required"`
	Default     jsontypes.Normalized            `tfsdk:"default"`
	Options     TypedList[jsontypes.Normalized] `tfsdk:"options"`
	Labels      *extensionParameterLabelsV0     `tfsdk:"labels"`
}

type extensionParameterLabelsV0 struct {
	Empty *string `tfsdk:"empty"`
	True  *string `tfsdk:"true"`
	False *string `tfsdk:"false"`
}

func (prior extensionStateV0) currentModel() ExtensionModel {
	model := ExtensionModel{
		IDIdentityModel: IDIdentityModel{ID: prior.ID},
		ExtensionIdentityModel: ExtensionIdentityModel{
			SpaceID: prior.SpaceID, EnvironmentID: prior.EnvironmentID, ExtensionID: prior.ExtensionID,
		},
		Parameters: prior.Parameters,
		Timeouts:   timeouts.Value{Object: prior.Timeouts},
	}
	if prior.Extension == nil {
		return model
	}

	model.Extension = &ExtensionConfiguration{
		Name: prior.Extension.Name, Src: prior.Extension.Src, SrcDoc: prior.Extension.SrcDoc, Sidebar: prior.Extension.Sidebar,
	}
	if prior.Extension.FieldTypes != nil {
		model.Extension.FieldTypes = make([]AppDefinitionLocationFieldTypesItem, len(prior.Extension.FieldTypes))
		for i, fieldType := range prior.Extension.FieldTypes {
			model.Extension.FieldTypes[i] = AppDefinitionLocationFieldTypesItem{
				Type: fieldType.Type, LinkType: fieldType.LinkType,
				Items: (*AppDefinitionLocationFieldTypeItemsItem)(fieldType.Items),
			}
		}
	}

	if prior.Extension.Parameters != nil {
		model.Extension.Parameters = &AppDefinitionParameters{
			Installation: extensionParametersV0ToCurrent(prior.Extension.Parameters.Installation),
			Instance:     extensionParametersV0ToCurrent(prior.Extension.Parameters.Instance),
		}
	}

	return model
}

func extensionParametersV0ToCurrent(prior []extensionParameterV0) []AppDefinitionParameter {
	if prior == nil {
		return nil
	}

	parameters := make([]AppDefinitionParameter, len(prior))
	for i, parameter := range prior {
		parameters[i] = AppDefinitionParameter{
			ID: parameter.ID, Type: parameter.Type, Name: parameter.Name,
			Description: parameter.Description, Required: parameter.Required,
			Default: parameter.Default, Options: parameter.Options,
			Labels: (*AppDefinitionParameterLabels)(parameter.Labels),
		}
	}

	return parameters
}
