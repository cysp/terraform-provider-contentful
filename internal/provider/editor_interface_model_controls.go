//nolint:dupl
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

func (v EditorInterfaceControlValue) ToEditorInterfaceDataControlsItem(_ context.Context, valuePath path.Path) (cm.EditorInterfaceDataControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldID, fieldIDDiags := requestRequiredString(v.FieldID, valuePath.AtName("field_id"))
	diags.Append(fieldIDDiags...)

	widgetNamespace, widgetNamespaceDiags := requestOptionalString(v.WidgetNamespace, valuePath.AtName("widget_namespace"))
	diags.Append(widgetNamespaceDiags...)

	widgetID, widgetIDDiags := requestOptionalString(v.WidgetID, valuePath.AtName("widget_id"))
	diags.Append(widgetIDDiags...)

	settings, settingsDiags := editorInterfaceOptionalSettings(v.Settings, valuePath.AtName("settings"))
	diags.Append(settingsDiags...)

	if diags.HasError() {
		return cm.EditorInterfaceDataControlsItem{}, diags
	}

	return cm.EditorInterfaceDataControlsItem{
		FieldId:         fieldID,
		WidgetNamespace: widgetNamespace,
		WidgetId:        widgetID,
		Settings:        settings,
	}, diags
}

func NewEditorInterfaceControlListValueFromResponse(_ context.Context, path path.Path, controlsItems []cm.EditorInterfaceControlsItem) (TypedList[TypedObject[EditorInterfaceControlValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceControlValue], len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		controlValue, controlValueDiags := NewEditorInterfaceControlValueFromResponse(path, item)
		diags.Append(controlValueDiags...)

		listElementValues[index] = controlValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceControlValueFromResponse(path path.Path, item cm.EditorInterfaceControlsItem) (TypedObject[EditorInterfaceControlValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceControlValue{
		FieldID:         types.StringValue(item.FieldId),
		WidgetNamespace: util.OptStringToStringValue(item.WidgetNamespace),
		WidgetID:        util.OptStringToStringValue(item.WidgetId),
		Settings:        jsontypes.NewNormalizedNull(),
	}

	if item.Settings != nil {
		settings, settingsErr := util.JxNormalizeOpaqueBytes(item.Settings, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if settingsErr != nil {
			diags.AddAttributeError(path.AtName("settings"), "Failed to read settings", settingsErr.Error())
		}

		value.Settings = NewNormalizedJSONTypesNormalizedValue(settings)
	}

	return NewTypedObject(value), diags
}
