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

func NewEditorsListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorsValue{}.Type(ctx))
}

func NewEditorsValueKnown() EditorsValue {
	return EditorsValue{
		state: attr.ValueStateKnown,
	}
}

func (model *EditorsValue) ToPutEditorInterfaceReqEditorsItem(_ context.Context, _ path.Path) (contentfulManagement.PutEditorInterfaceReqEditorsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqEditorsItem{
		WidgetNamespace: model.WidgetNamespace.ValueString(),
		WidgetId:        model.WidgetId.ValueString(),
		Disabled:        util.BoolValueToOptBool(model.Disabled),
		Settings:        []byte(model.Settings.ValueString()),
	}

	return item, diags
}

func NewEditorsListValueFromResponse(ctx context.Context, path path.Path, editors []contentfulManagement.EditorInterfaceEditorsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	editorsListValues := make([]attr.Value, len(editors))

	for index, item := range editors {
		path := path.AtListIndex(index)

		editorsValue, editorsValueDiags := NewEditorsValueFromResponse(path, item)
		diags.Append(editorsValueDiags...)

		editorsListValues[index] = editorsValue
	}

	list, listDiags := types.ListValue(EditorsValue{}.Type(ctx), editorsListValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewEditorsValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceEditorsItem) (EditorsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorsValue{
		WidgetNamespace: types.StringValue(item.WidgetNamespace),
		WidgetId:        types.StringValue(item.WidgetId),
		Disabled:        util.OptBoolToBoolValue(item.Disabled),
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
