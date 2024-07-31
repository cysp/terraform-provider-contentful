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

func (model *ControlsValue) ToPutEditorInterfaceReqControlsItem(ctx context.Context, path path.Path) (contentfulManagement.PutEditorInterfaceReqControlsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqControlsItem{
		FieldId:         model.FieldId.ValueString(),
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

func NewControlsValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceControlsItem) (ControlsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

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

	return value, diags
}

func NewControlsListValueNull(ctx context.Context) types.List {
	return types.ListNull(ControlsValue{}.Type(ctx))
}

func NewControlsListValueFromResponse(ctx context.Context, path path.Path, controlsItems []contentfulManagement.EditorInterfaceControlsItem) (types.List, diag.Diagnostics) {
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
