//nolint:revive,stylecheck
package provider

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

func NewEditorsValueKnown() EditorsValue {
	return EditorsValue{
		state: attr.ValueStateKnown,
	}
}

func (model *EditorsValue) ToPutEditorInterfaceReqEditorsItem(ctx context.Context, path path.Path) (contentfulManagement.PutEditorInterfaceReqEditorsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqEditorsItem{
		WidgetNamespace: model.WidgetNamespace.ValueString(),
		WidgetId:        model.WidgetId.ValueString(),
		Disabled:        util.BoolValueToOptBool(model.Disabled),
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

func NewEditorsValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceEditorsItem) (EditorsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorsValue{
		WidgetNamespace: types.StringValue(item.WidgetNamespace),
		WidgetId:        types.StringValue(item.WidgetId),
		Disabled:        util.OptBoolToBoolValue(item.Disabled),
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

func NewEditorsListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorsValue{}.Type(ctx))
}

func NewEditorsListValueFromResponse(ctx context.Context, path path.Path, controlsItems []contentfulManagement.EditorInterfaceEditorsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		EditorsValue, EditorsValueDiags := NewEditorsValueFromResponse(path, item)
		diags.Append(EditorsValueDiags...)

		listElementValues[index] = EditorsValue
	}

	list, listDiags := types.ListValue(EditorsValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
