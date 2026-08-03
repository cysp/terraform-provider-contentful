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

func (v EditorInterfaceGroupControlValue) ToEditorInterfaceDataGroupControlsItem(_ context.Context, valuePath path.Path) (cm.EditorInterfaceDataGroupControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	groupID, groupIDDiags := requestRequiredString(v.GroupID, valuePath.AtName("group_id"))
	diags.Append(groupIDDiags...)

	widgetNamespace, widgetNamespaceDiags := requestOptionalString(v.WidgetNamespace, valuePath.AtName("widget_namespace"))
	diags.Append(widgetNamespaceDiags...)

	widgetID, widgetIDDiags := requestOptionalString(v.WidgetID, valuePath.AtName("widget_id"))
	diags.Append(widgetIDDiags...)

	settings, settingsDiags := editorInterfaceOptionalSettings(v.Settings, valuePath.AtName("settings"))
	diags.Append(settingsDiags...)

	if diags.HasError() {
		return cm.EditorInterfaceDataGroupControlsItem{}, diags
	}

	return cm.EditorInterfaceDataGroupControlsItem{
		GroupId:         groupID,
		WidgetNamespace: widgetNamespace,
		WidgetId:        widgetID,
		Settings:        settings,
	}, diags
}

func NewEditorInterfaceGroupControlListValueFromResponse(_ context.Context, path path.Path, groupControlsItems []cm.EditorInterfaceGroupControlsItem) (TypedList[TypedObject[EditorInterfaceGroupControlValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceGroupControlValue], len(groupControlsItems))

	for index, item := range groupControlsItems {
		path := path.AtListIndex(index)

		groupControlValue, groupControlValueDiags := NewEditorInterfaceGroupControlValueFromResponse(path, item)
		diags.Append(groupControlValueDiags...)

		listElementValues[index] = groupControlValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceGroupControlValueFromResponse(path path.Path, item cm.EditorInterfaceGroupControlsItem) (TypedObject[EditorInterfaceGroupControlValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceGroupControlValue{
		GroupID:         types.StringValue(item.GroupId),
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
