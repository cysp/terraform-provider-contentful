//nolint:dupl,revive,stylecheck
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

func NewGroupControlsValueKnown() GroupControlsValue {
	return GroupControlsValue{
		state: attr.ValueStateKnown,
	}
}

func (model *GroupControlsValue) ToPutEditorInterfaceReqGroupControlsItem(ctx context.Context, path path.Path) (contentfulManagement.PutEditorInterfaceReqGroupControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqGroupControlsItem{
		GroupId:         model.GroupId.ValueString(),
		WidgetNamespace: util.StringValueToOptString(model.WidgetNamespace),
		WidgetId:        util.StringValueToOptString(model.WidgetId),
	}

	if model.Settings.IsNull() || model.Settings.IsUnknown() {
	} else {
		modelSettings := model.Settings.ValueString()

		path := path.AtName("settings")

		if modelSettings != "" {
			decoder := jx.DecodeStr(modelSettings)

			err := item.Settings.Decode(decoder)
			if err != nil {
				diags.AddAttributeError(path, "Failed to decode settings", err.Error())
			}
		}
	}

	return item, diags
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

	if settings, ok := item.Settings.Get(); ok {
		encoder := jx.Encoder{}
		util.EncodeJxRawMapOrdered(&encoder, settings)
		value.Settings = types.StringValue(encoder.String())
	}

	return value, diags
}

func NewGroupControlsListValueNull(ctx context.Context) types.List {
	return types.ListNull(GroupControlsValue{}.Type(ctx))
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
