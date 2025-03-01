//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewControlsListValueNull(ctx context.Context) types.List {
	return types.ListNull(ControlsValue{}.Type(ctx))
}

func NewControlsValueKnown() ControlsValue {
	return ControlsValue{
		state: attr.ValueStateKnown,
	}
}

func (model *ControlsValue) ToEditorInterfaceFieldsControlsItem(_ context.Context, _ path.Path) (cm.EditorInterfaceFieldsControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceFieldsControlsItem{
		FieldId:         model.FieldId.ValueString(),
		WidgetNamespace: util.StringValueToOptString(model.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(model.WidgetId),
	}

	modelSettingsString := model.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewControlsListValueFromResponse(ctx context.Context, path path.Path, controlsItems []cm.EditorInterfaceControlsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		controlsValue, controlsValueDiags := NewControlsValueFromResponse(path, item)
		diags.Append(controlsValueDiags...)

		listElementValues[index] = controlsValue
	}

	list, listDiags := types.ListValue(ControlsValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewControlsValueFromResponse(path path.Path, item cm.EditorInterfaceControlsItem) (ControlsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ControlsValue{
		FieldId:         types.StringValue(item.FieldId),
		WidgetNamespace: util.OptStringToStringValue(item.WidgetNamespace),
		WidgetId:        util.OptStringToStringValue(item.WidgetId),
		Settings:        types.StringNull(),
		state:           attr.ValueStateKnown,
	}

	if item.Settings != nil {
		settings, settingsErr := util.JxNormalizeOpaqueBytes(item.Settings, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if settingsErr != nil {
			diags.AddAttributeError(path.AtName("settings"), "Failed to read settings", settingsErr.Error())
		}

		value.Settings = types.StringValue(string(settings))
	}

	return value, diags
}
