package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewExtensionModelFromResponse(ctx context.Context, response cm.Extension) (ExtensionModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := response.Sys.Space.Sys.ID
	environmentID := response.Sys.Environment.Sys.ID
	extensionID := response.Sys.ID

	model := ExtensionModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, extensionID),
		ExtensionIdentityModel: ExtensionIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			ExtensionID:   types.StringValue(extensionID),
		},
	}

	extensionExtension, extensionExtensionDiags := NewExtensionModelExtensionFromResponse(ctx, response.Extension)
	diags.Append(extensionExtensionDiags...)

	model.Extension = &extensionExtension

	if response.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(response.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = NewNormalizedJSONTypesNormalizedValue(constraint)
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	return model, diags
}

func NewExtensionModelExtensionFromResponse(ctx context.Context, response cm.ExtensionExtension) (ExtensionModelExtension, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := ExtensionModelExtension{
		Name:    types.StringValue(response.Name),
		Src:     types.StringValue(response.Src.Or("")),
		SrcDoc:  types.StringValue(response.Srcdoc.Or("")),
		Sidebar: types.BoolPointerValue(response.Sidebar.ValueBoolPointer()),
	}

	if response.FieldTypes != nil {
		fieldTypes := make([]AppDefinitionLocationFieldTypesItem, 0, len(response.FieldTypes))

		for _, fieldType := range response.FieldTypes {
			fieldTypeItem := AppDefinitionLocationFieldTypesItem{
				Type:     types.StringValue(fieldType.Type),
				LinkType: types.StringPointerValue(fieldType.LinkType.ValueStringPointer()),
			}

			if fieldTypeItems, ok := fieldType.Items.Get(); ok {
				fieldTypeItem.Items = &AppDefinitionLocationFieldTypeItemsItem{
					Type:     types.StringValue(fieldTypeItems.Type),
					LinkType: types.StringPointerValue(fieldTypeItems.LinkType.ValueStringPointer()),
				}
			}

			fieldTypes = append(fieldTypes, fieldTypeItem)
		}

		model.FieldTypes = fieldTypes
	}

	if parameters, ok := response.Parameters.Get(); ok {
		var installationParameters, instanceParameters []AppDefinitionParameter

		if parameters.Installation != nil {
			installationParameters = make([]AppDefinitionParameter, 0, len(parameters.Installation))

			for _, element := range parameters.Installation {
				parameter, parameterDiags := NewAppDefinitionParameterFromResponse(ctx, element)
				diags.Append(parameterDiags...)

				installationParameters = append(installationParameters, parameter)
			}
		}

		if parameters.Instance != nil {
			instanceParameters = make([]AppDefinitionParameter, 0, len(parameters.Instance))

			for _, element := range parameters.Instance {
				parameter, parameterDiags := NewAppDefinitionParameterFromResponse(ctx, element)
				diags.Append(parameterDiags...)

				instanceParameters = append(instanceParameters, parameter)
			}
		}

		model.Parameters = &AppDefinitionParameters{
			Installation: installationParameters,
			Instance:     instanceParameters,
		}
	}

	return model, diags
}
