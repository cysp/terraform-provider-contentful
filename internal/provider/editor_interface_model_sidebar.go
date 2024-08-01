package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSidebarValueKnown() SidebarValue {
	return SidebarValue{
		state: attr.ValueStateKnown,
	}
}

func (model *SidebarValue) ToPutEditorInterfaceReqSidebarItem(_ context.Context, _ path.Path) (contentfulManagement.PutEditorInterfaceReqSidebarItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqSidebarItem{
		WidgetNamespace: model.WidgetNamespace.ValueString(),
		WidgetId:        model.WidgetId.ValueString(),
	}

	modelDisabled := model.Disabled.ValueBoolPointer()
	if modelDisabled != nil {
		item.Disabled.SetTo(*modelDisabled)
	}

	modelSettingsString := model.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewSidebarValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceSidebarItem) (SidebarValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := SidebarValue{
		WidgetNamespace: types.StringValue(item.WidgetNamespace),
		WidgetId:        types.StringValue(item.WidgetId),
		Disabled:        types.BoolNull(),
		Settings:        types.StringNull(),
		state:           attr.ValueStateKnown,
	}

	if disabled, ok := item.Disabled.Get(); ok {
		value.Disabled = types.BoolValue(disabled)
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

func NewSidebarListValueNull(ctx context.Context) types.List {
	return types.ListNull(SidebarValue{}.Type(ctx))
}

func NewSidebarListValueFromResponse(ctx context.Context, path path.Path, sidebarItems []contentfulManagement.EditorInterfaceSidebarItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(sidebarItems))

	for index, item := range sidebarItems {
		sidebarValue, sidebarValueDiags := NewSidebarValueFromResponse(path.AtListIndex(index), item)
		diags.Append(sidebarValueDiags...)

		listElementValues[index] = sidebarValue
	}

	list, listDiags := types.ListValue(SidebarValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
