package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *ExtensionModel) validateRequestConfiguration(config ExtensionModel, modelPath path.Path) diag.Diagnostics {
	diags := diag.Diagnostics{}
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(model.Parameters, config.Parameters, modelPath.AtName("parameters"))...)

	if model.Extension != nil {
		extensionPath := modelPath.AtName("extension")

		var configured ExtensionModelExtension

		if config.Extension != nil {
			configured = *config.Extension
		}

		diags.Append(rejectUnknownConfigurationOwnedRequestValue(model.Extension.Src, configured.Src, extensionPath.AtName("src"))...)
		diags.Append(rejectUnknownConfigurationOwnedRequestValue(model.Extension.SrcDoc, configured.SrcDoc, extensionPath.AtName("srcdoc"))...)
	}

	return diags
}

func (model *ExtensionModel) ToExtensionData(config ExtensionModel, path path.Path) (cm.ExtensionData, diag.Diagnostics) {
	diags := model.validateRequestConfiguration(config, path)
	if diags.HasError() {
		return cm.ExtensionData{}, diags
	}

	fields := cm.ExtensionData{}

	if model.Extension == nil {
		diags.AddAttributeError(
			path.AtName("extension"),
			"Unexpected null extension configuration",
			"The extension configuration is required before an extension request can be sent to Contentful.",
		)

		return cm.ExtensionData{}, diags
	}

	sourcePath := path.AtName("extension").AtName("src")
	srcPresent := !model.Extension.Src.IsNull() && !model.Extension.Src.IsUnknown()
	srcdocPresent := !model.Extension.SrcDoc.IsNull() && !model.Extension.SrcDoc.IsUnknown()

	switch {
	case srcPresent && srcdocPresent:
		diags.AddAttributeError(
			path.AtName("extension").AtName("srcdoc"),
			"Conflicting extension sources",
			"A Contentful extension request must contain exactly one of extension.src or extension.srcdoc.",
		)
	case !srcPresent && !srcdocPresent:
		if model.Extension.SrcDoc.IsUnknown() {
			sourcePath = path.AtName("extension").AtName("srcdoc")
		}

		diags.AddAttributeError(
			sourcePath,
			"Missing extension source",
			"A Contentful extension request must contain exactly one known, non-null extension.src or extension.srcdoc value.",
		)
	}

	fieldsExtension, fieldsExtensionDiags := model.Extension.ToExtensionExtensionData(path.AtName("extension"))
	diags.Append(fieldsExtensionDiags...)

	fields.Extension = fieldsExtension

	// An unknown Optional+Computed plan is omitted only when configuration is
	// null. Every known plan value remains the request source.
	if !model.Parameters.IsUnknown() && !model.Parameters.IsNull() {
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	if diags.HasError() {
		return cm.ExtensionData{}, diags
	}

	return fields, diags
}

func (model *ExtensionModelExtension) ToExtensionExtensionData(
	path path.Path,
) (cm.ExtensionDataExtension, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, path.AtName("name"))
	diags.Append(nameDiags...)

	var sidebar *bool

	switch {
	case model.Sidebar.IsUnknown():
		diags.AddAttributeError(
			path.AtName("sidebar"),
			"Unexpected unknown request value",
			"This value must be known before it can be sent to Contentful.",
		)
	case model.Sidebar.IsNull():
	default:
		sidebar = model.Sidebar.ValueBoolPointer()
	}

	fields := cm.ExtensionDataExtension{
		Name:    name,
		Sidebar: cm.NewOptPointerBool(sidebar),
	}

	if !model.Src.IsUnknown() && !model.Src.IsNull() {
		fields.Src = cm.NewOptString(model.Src.ValueString())
	}

	if !model.SrcDoc.IsUnknown() && !model.SrcDoc.IsNull() {
		fields.Srcdoc = cm.NewOptString(model.SrcDoc.ValueString())
	}

	if model.FieldTypes != nil {
		fieldTypes := make([]cm.ExtensionDataExtensionFieldTypesItem, 0, len(model.FieldTypes))

		for index, fieldType := range model.FieldTypes {
			requestFieldType, fieldTypeDiags := appDefinitionLocationFieldTypeToExtensionDataExtensionFieldTypesItem(
				fieldType,
				path.AtName("field_types").AtListIndex(index),
			)
			diags.Append(fieldTypeDiags...)

			if !fieldTypeDiags.HasError() {
				fieldTypes = append(fieldTypes, requestFieldType)
			}
		}

		fields.FieldTypes = fieldTypes
	}

	if model.Parameters != nil {
		parameters, parameterDiags := model.Parameters.ToAppDefinitionParameters(path.AtName("parameters"))
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

func appDefinitionLocationFieldTypeToExtensionDataExtensionFieldTypesItem(
	fieldType AppDefinitionLocationFieldTypesItem,
	path path.Path,
) (cm.ExtensionDataExtensionFieldTypesItem, diag.Diagnostics) {
	requestFieldType, diags := appDefinitionLocationFieldTypeToRequest(fieldType, path)

	if diags.HasError() {
		return cm.ExtensionDataExtensionFieldTypesItem{}, diags
	}

	fieldTypesItem := cm.ExtensionDataExtensionFieldTypesItem{
		Type:     requestFieldType.Type,
		LinkType: cm.NewOptPointerString(requestFieldType.LinkType),
	}

	if requestFieldType.Items != nil {
		fieldTypesItem.Items.SetTo(cm.ExtensionDataExtensionFieldTypesItemItems{
			Type:     requestFieldType.Items.Type,
			LinkType: cm.NewOptPointerString(requestFieldType.Items.LinkType),
		})
	}

	return fieldTypesItem, diags
}
