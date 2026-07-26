package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *ExtensionModel) ToExtensionData(ctx context.Context, path path.Path) (cm.ExtensionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.ExtensionData{}

	fieldsExtension, fieldsExtensionDiags := model.Extension.ToExtensionExtensionData(ctx, path.AtName("extension"))
	diags.Append(fieldsExtensionDiags...)

	fields.Extension = fieldsExtension

	if !model.Parameters.IsUnknown() && !model.Parameters.IsNull() {
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	if diags.HasError() {
		return cm.ExtensionData{}, diags
	}

	return fields, diags
}

func (model *ExtensionModelExtension) ToExtensionExtensionData(ctx context.Context, path path.Path) (cm.ExtensionDataExtension, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.ExtensionDataExtension{
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
		fieldTypes := make([]cm.ExtensionDataExtensionFieldTypesItem, 0, len(model.FieldTypes))

		for _, fieldType := range model.FieldTypes {
			fieldTypes = append(fieldTypes, AppDefinitionLocationFieldTypesItemToExtensionDataExtensionFieldTypesItem(fieldType))
		}

		fields.FieldTypes = fieldTypes
	}

	if model.Parameters != nil {
		parameters, parameterDiags := toAppDefinitionParameters(ctx, path.AtName("parameters"), *model.Parameters)
		diags.Append(parameterDiags...)

		if !parameterDiags.HasError() {
			fields.Parameters.SetTo(parameters)
		}
	}

	if diags.HasError() {
		return cm.ExtensionDataExtension{}, diags
	}

	return fields, diags
}

func AppDefinitionLocationFieldTypesItemToExtensionDataExtensionFieldTypesItem(
	fieldType AppDefinitionLocationFieldTypesItem,
) cm.ExtensionDataExtensionFieldTypesItem {
	fieldTypesItem := cm.ExtensionDataExtensionFieldTypesItem{
		Type:     fieldType.Type.ValueString(),
		LinkType: cm.NewOptPointerString(fieldType.LinkType.ValueStringPointer()),
	}

	if fieldType.Items != nil {
		fieldTypesItem.Items.SetTo(cm.ExtensionDataExtensionFieldTypesItemItems{
			Type:     fieldType.Items.Type.ValueString(),
			LinkType: cm.NewOptPointerString(fieldType.Items.LinkType.ValueStringPointer()),
		})
	}

	return fieldTypesItem
}
