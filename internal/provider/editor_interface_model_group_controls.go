//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceGroupControlListValueNull() TypedList[EditorInterfaceGroupControlValue] {
	return NewTypedListNull[EditorInterfaceGroupControlValue]()
}

func NewEditorInterfaceGroupControlValueKnown() EditorInterfaceGroupControlValue {
	return EditorInterfaceGroupControlValue{
		state: attr.ValueStateKnown,
	}
}

func (v *EditorInterfaceGroupControlValue) ToEditorInterfaceFieldsGroupControlsItem(_ context.Context, _ path.Path) (cm.EditorInterfaceFieldsGroupControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceFieldsGroupControlsItem{
		GroupId:         v.GroupID.ValueString(),
		WidgetNamespace: util.StringValueToOptString(v.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(v.WidgetID),
	}

	modelSettingsString := v.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewEditorInterfaceGroupControlListValueFromResponse(_ context.Context, path path.Path, groupControlsItems []cm.EditorInterfaceGroupControlsItem) (TypedList[EditorInterfaceGroupControlValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]EditorInterfaceGroupControlValue, len(groupControlsItems))

	for index, item := range groupControlsItems {
		path := path.AtListIndex(index)

		groupControlValue, groupControlValueDiags := NewEditorInterfaceGroupControlValueFromResponse(path, item)
		diags.Append(groupControlValueDiags...)

		listElementValues[index] = groupControlValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceGroupControlValueFromResponse(path path.Path, item cm.EditorInterfaceGroupControlsItem) (EditorInterfaceGroupControlValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceGroupControlValue{
		GroupID:         types.StringValue(item.GroupId),
		WidgetNamespace: util.OptStringToStringValue(item.WidgetNamespace),
		WidgetID:        util.OptStringToStringValue(item.WidgetId),
		Settings:        jsontypes.NewNormalizedNull(),
		state:           attr.ValueStateKnown,
	}

	if item.Settings != nil {
		settings, settingsErr := util.JxNormalizeOpaqueBytes(item.Settings, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if settingsErr != nil {
			diags.AddAttributeError(path.AtName("settings"), "Failed to read settings", settingsErr.Error())
		}

		value.Settings = jsontypes.NewNormalizedValue(string(settings))
	}

	return value, diags
}
