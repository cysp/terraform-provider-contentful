//nolint:dupl
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

func NewGroupControlsListValueNull(ctx context.Context) types.List {
	return types.ListNull(GroupControlsValue{}.Type(ctx))
}

func NewGroupControlsValueKnown() GroupControlsValue {
	return GroupControlsValue{
		state: attr.ValueStateKnown,
	}
}

func (model *GroupControlsValue) ToPutEditorInterfaceReqGroupControlsItem(_ context.Context, _ path.Path) (contentfulManagement.PutEditorInterfaceReqGroupControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqGroupControlsItem{
		GroupId:         model.GroupId.ValueString(),
		WidgetNamespace: util.StringValueToOptString(model.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(model.WidgetId),
	}

	modelSettingsString := model.Settings.ValueString()
	if modelSettingsString != "" {
		item.Settings = []byte(modelSettingsString)
	}

	return item, diags
}

func NewGroupControlsListValueFromResponse(ctx context.Context, path path.Path, groupControlsItems []contentfulManagement.EditorInterfaceGroupControlsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(groupControlsItems))

	for index, item := range groupControlsItems {
		path := path.AtListIndex(index)

		groupControlsValue, groupControlsValueDiags := NewGroupControlsValueFromResponse(path, item)
		diags.Append(groupControlsValueDiags...)

		listElementValues[index] = groupControlsValue
	}

	list, listDiags := types.ListValue(GroupControlsValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewGroupControlsValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceGroupControlsItem) (GroupControlsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := GroupControlsValue{
		GroupId:         types.StringValue(item.GroupId),
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
