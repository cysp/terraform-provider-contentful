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

func NewEditorInterfaceControlListValueNull(ctx context.Context) TypedList[EditorInterfaceControlValue] {
	return NewTypedListNull[EditorInterfaceControlValue](ctx)
}

func NewEditorInterfaceControlValueKnown() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		state: attr.ValueStateKnown,
	}
}

func (v *EditorInterfaceControlValue) ToEditorInterfaceFieldsControlsItem(_ context.Context, _ path.Path) (cm.EditorInterfaceFieldsControlsItem, diag.Diagnostics) {
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

func NewEditorInterfaceControlListValueFromResponse(ctx context.Context, path path.Path, controlsItems []cm.EditorInterfaceControlsItem) (TypedList[EditorInterfaceControlValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]EditorInterfaceControlValue, len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		controlValue, controlValueDiags := NewEditorInterfaceControlValueFromResponse(path, item)
		diags.Append(controlValueDiags...)

		listElementValues[index] = controlValue
	}

	list, listDiags := NewTypedList(ctx, listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewEditorInterfaceControlValueFromResponse(path path.Path, item cm.EditorInterfaceControlsItem) (EditorInterfaceControlValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceControlValue{
		FieldID:         types.StringValue(item.FieldId),
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
