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

func (v EditorInterfaceSidebarValue) ToEditorInterfaceDataSidebarItem(_ context.Context, _ path.Path) (cm.EditorInterfaceDataSidebarItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceDataSidebarItem{
		WidgetNamespace: v.WidgetNamespace.ValueString(),
		WidgetId:        v.WidgetID.ValueString(),
	}

	modelDisabled := v.Disabled.ValueBoolPointer()
	if modelDisabled != nil {
		item.Disabled.SetTo(*modelDisabled)
	}

	modelSettingsString := v.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewEditorInterfaceSidebarValueFromResponse(path path.Path, item cm.EditorInterfaceSidebarItem) (TypedObject[EditorInterfaceSidebarValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceSidebarValue{
		WidgetNamespace: types.StringValue(item.WidgetNamespace),
		WidgetID:        types.StringValue(item.WidgetId),
		Disabled:        types.BoolNull(),
		Settings:        jsontypes.NewNormalizedNull(),
	}

	if disabled, ok := item.Disabled.Get(); ok {
		value.Disabled = types.BoolValue(disabled)
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

func NewEditorInterfaceSidebarListValueFromResponse(_ context.Context, path path.Path, sidebarItems []cm.EditorInterfaceSidebarItem) (TypedList[TypedObject[EditorInterfaceSidebarValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceSidebarValue], len(sidebarItems))

	for index, item := range sidebarItems {
		sidebarValue, sidebarValueDiags := NewEditorInterfaceSidebarValueFromResponse(path.AtListIndex(index), item)
		diags.Append(sidebarValueDiags...)

		listElementValues[index] = sidebarValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}
