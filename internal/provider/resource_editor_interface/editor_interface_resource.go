//nolint:revive,stylecheck
package resource_editor_interface

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewControlsValueKnown() ControlsValue {
	return ControlsValue{
		state: attr.ValueStateKnown,
	}
}

func NewSidebarValueKnown() SidebarValue {
	return SidebarValue{
		state: attr.ValueStateKnown,
	}
}

func (model EditorInterfaceModel) ToPutEditorInterfaceReq(ctx context.Context, diags *diag.Diagnostics) contentfulManagement.PutEditorInterfaceReq {
	request := contentfulManagement.PutEditorInterfaceReq{}

	controlsPath := path.Root("controls")

	controlsElementValues := []ControlsValue{}
	diags.Append(model.Controls.ElementsAs(ctx, &controlsElementValues, false)...)

	request.Controls = make([]contentfulManagement.PutEditorInterfaceReqControlsItem, len(controlsElementValues))

	for index, controlsElement := range controlsElementValues {
		path := controlsPath.AtListIndex(index)
		request.Controls[index] = controlsElement.ToPutEditorInterfaceReqControlsItem(ctx, path, diags)
	}

	sidebarPath := path.Root("sidebar")

	sidebarElementValues := []SidebarValue{}
	diags.Append(model.Sidebar.ElementsAs(ctx, &sidebarElementValues, false)...)

	request.Sidebar = make([]contentfulManagement.PutEditorInterfaceReqSidebarItem, len(sidebarElementValues))

	for index, sidebarElement := range sidebarElementValues {
		path := sidebarPath.AtListIndex(index)
		request.Sidebar[index] = sidebarElement.ToPutEditorInterfaceReqSidebarItem(ctx, path, diags)
	}

	return request
}

func (model ControlsValue) ToPutEditorInterfaceReqControlsItem(ctx context.Context, path path.Path, diags *diag.Diagnostics) contentfulManagement.PutEditorInterfaceReqControlsItem {
	item := contentfulManagement.PutEditorInterfaceReqControlsItem{
		FieldId:         model.FieldId.ValueString(),
		WidgetNamespace: util.StringValueToOptString(model.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(model.WidgetId),
	}

	modelSettings := model.Settings.ValueStringPointer()
	if modelSettings != nil {
		path := path.AtName("settings")

		decoder := jx.DecodeStr(*modelSettings)

		err := item.Settings.Decode(decoder)
		if err != nil {
			diags.AddAttributeError(path, "Failed to decode settings", err.Error())
		}
	}

	return item
}

func (model SidebarValue) ToPutEditorInterfaceReqSidebarItem(ctx context.Context, path path.Path, diags *diag.Diagnostics) contentfulManagement.PutEditorInterfaceReqSidebarItem {
	item := contentfulManagement.PutEditorInterfaceReqSidebarItem{
		WidgetNamespace: model.WidgetNamespace.ValueString(),
		WidgetId:        model.WidgetId.ValueString(),
	}

	modelDisabled := model.Disabled.ValueBoolPointer()
	if modelDisabled != nil {
		item.Disabled.SetTo(*modelDisabled)
	}

	modelSettings := model.Settings.ValueStringPointer()
	if modelSettings != nil {
		path := path.AtName("settings")

		decoder := jx.DecodeStr(*modelSettings)

		err := item.Settings.Decode(decoder)
		if err != nil {
			diags.AddAttributeError(path, "Failed to decode settings", err.Error())
		}
	}

	return item
}

func NewControlsValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceControlsItem) ControlsValue {
	value := ControlsValue{
		FieldId:         types.StringValue(item.FieldId),
		WidgetNamespace: util.OptStringToStringValue(item.WidgetNamespace),
		WidgetId:        util.OptStringToStringValue(item.WidgetId),
		Settings:        types.StringNull(),
		state:           attr.ValueStateKnown,
	}

	if settings, ok := item.Settings.Get(); ok {
		encoder := jx.Encoder{}
		util.EncodeJxRawMapOrdered(&encoder, settings)
		value.Settings = types.StringValue(encoder.String())
	}

	return value
}

func NewControlsListValueFromResponse(ctx context.Context, path path.Path, controlsItems []contentfulManagement.EditorInterfaceControlsItem, diags *diag.Diagnostics) types.List {
	listElementValues := make([]attr.Value, len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		listElementValues[index] = NewControlsValueFromResponse(path, item)
	}

	list, listDiags := types.ListValue(ControlsValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list
}

func NewSidebarValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceSidebarItem) SidebarValue {
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

	if settings, ok := item.Settings.Get(); ok {
		encoder := jx.Encoder{}
		util.EncodeJxRawMapOrdered(&encoder, settings)
		value.Settings = types.StringValue(encoder.String())
	}

	return value
}

func NewSidebarListValueFromResponse(ctx context.Context, path path.Path, sidebarItems []contentfulManagement.EditorInterfaceSidebarItem, diags *diag.Diagnostics) types.List {
	listElementValues := make([]attr.Value, len(sidebarItems))

	for index, item := range sidebarItems {
		path := path.AtListIndex(index)
		listElementValues[index] = NewSidebarValueFromResponse(path, item)
	}

	list, listDiags := types.ListValue(SidebarValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list
}
