package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *ExtensionModel) ToExtensionFields(ctx context.Context, path path.Path) (cm.ExtensionFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.ExtensionFields{}

	fieldsExtension, fieldsExtensionDiags := model.Extension.ToExtensionExtensionFields(ctx, path.AtName("extension"))
	diags.Append(fieldsExtensionDiags...)

	fields.Extension = fieldsExtension

	if !model.Parameters.IsUnknown() && !model.Parameters.IsNull() {
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	return fields, diags
}

func (model *ExtensionModelExtension) ToExtensionExtensionFields(ctx context.Context, path path.Path) (cm.ExtensionFieldsExtension, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.ExtensionFieldsExtension{
		Name:    model.Name.ValueString(),
		Sidebar: cm.NewOptPointerBool(model.Sidebar.ValueBoolPointer()),
	}

	src := model.Src.ValueString()
	if src != "" {
		fields.Src = cm.NewOptString(src)
	}

	srcdoc := model.SrcDoc.ValueString()
	if srcdoc != "" {
		fields.Srcdoc = cm.NewOptString(srcdoc)
	}

	if model.FieldTypes != nil {
		fieldTypes := make([]cm.ExtensionFieldsExtensionFieldTypesItem, 0, len(model.FieldTypes))

		for _, fieldType := range model.FieldTypes {
			fieldTypes = append(fieldTypes, AppDefinitionLocationFieldTypesItemToExtensionFieldsExtensionFieldTypesItem(fieldType))
		}

		fields.FieldTypes = fieldTypes
	}

	if model.Parameters != nil {
		path := path.AtName("parameters")

		parameters := cm.AppDefinitionParameters{}

		if model.Parameters.Installation != nil {
			path := path.AtName("installation")

			installationParameters := make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Installation))

			for index, parameter := range model.Parameters.Installation {
				path := path.AtListIndex(index)

				installationParameter, parameterDiags := parameter.ToAppDefinitionParameter(ctx, path)
				diags.Append(parameterDiags...)

				installationParameters = append(installationParameters, installationParameter)
			}

			parameters.Installation = installationParameters
		}

		if model.Parameters.Instance != nil {
			path := path.AtName("instance")

			instanceParameters := make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Instance))

			for index, parameter := range model.Parameters.Instance {
				path := path.AtListIndex(index)

				instanceParameter, parameterDiags := parameter.ToAppDefinitionParameter(ctx, path)
				diags.Append(parameterDiags...)

				instanceParameters = append(instanceParameters, instanceParameter)
			}

			parameters.Instance = instanceParameters
		}

		fields.Parameters.SetTo(parameters)
	}

	return fields, diags
}

func AppDefinitionLocationFieldTypesItemToExtensionFieldsExtensionFieldTypesItem(
	fieldType AppDefinitionLocationFieldTypesItem,
) cm.ExtensionFieldsExtensionFieldTypesItem {
	fieldTypesItem := cm.ExtensionFieldsExtensionFieldTypesItem{
		Type:     fieldType.Type.ValueString(),
		LinkType: cm.NewOptPointerString(fieldType.LinkType.ValueStringPointer()),
	}

	if fieldType.Items != nil {
		fieldTypesItem.Items.SetTo(cm.ExtensionFieldsExtensionFieldTypesItemItems{
			Type:     fieldType.Items.Type.ValueString(),
			LinkType: cm.NewOptPointerString(fieldType.Items.LinkType.ValueStringPointer()),
		})
	}

	return fieldTypesItem
}
