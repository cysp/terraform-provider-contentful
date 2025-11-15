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

func (v EditorInterfaceControlValue) ToEditorInterfaceFieldsControlsItem(_ context.Context, _ path.Path) (cm.EditorInterfaceFieldsControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceFieldsControlsItem{
		FieldId:         v.FieldID.ValueString(),
		WidgetNamespace: util.StringValueToOptString(v.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(v.WidgetID),
	}

	modelSettingsString := v.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewEditorInterfaceControlListValueFromResponse(_ context.Context, path path.Path, controlsItems []cm.EditorInterfaceControlsItem) (TypedList[TypedObject[EditorInterfaceControlValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceControlValue], len(controlsItems))

	for index, item := range controlsItems {
		controlPath := path.AtListIndex(index)

		controlValue, controlValueDiags := NewEditorInterfaceControlValueFromResponse(controlPath, item)
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

		value.Settings = jsontypes.NewNormalizedValue(string(settings))
	}

	return NewTypedObject(value), diags
}
