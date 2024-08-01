//nolint:revive,stylecheck
package resource_editor_interface

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorLayoutValueKnown() EditorLayoutValue {
	return EditorLayoutValue{
		state: attr.ValueStateKnown,
	}
}

func (model *EditorLayoutValue) ToPutEditorInterfaceReqEditorLayoutItem(ctx context.Context, path path.Path) (contentfulManagement.PutEditorInterfaceReqEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqEditorLayoutItem{
		GroupId: model.GroupId.ValueString(),
		Name:    model.Name.ValueString(),
	}

	if model.Items.IsNull() || model.Items.IsUnknown() {
	} else {
		// modelItems := model.Items.ValueString()

		// path := path.AtName("items")

		// if modelItems != "" {
		// 	decoder := jx.DecodeStr(modelItems)

		// 	err := item.Items.Decode(decoder)
		// 	if err != nil {
		// 		diags.AddAttributeError(path, "Failed to decode items", err.Error())
		// 	}
		// }
	}

	return item, diags
}

func NewEditorLayoutValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceEditorLayoutItem) (EditorLayoutValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorLayoutValue{
		GroupId: types.StringValue(item.GroupId),
		Name:    types.StringValue(item.Name),
		state:   attr.ValueStateKnown,
	}

	// if Items, ok := item.Items.Get(); ok {
	// 	encoder := jx.Encoder{}
	// 	util.EncodeJxRawMapOrdered(&encoder, Items)
	// 	value.Items = types.StringValue(encoder.String())
	// }

	return value, diags
}

func NewEditorLayoutListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorLayoutValue{}.Type(ctx))
}

func NewEditorLayoutListValueFromResponse(ctx context.Context, path path.Path, controlsItems []contentfulManagement.EditorInterfaceEditorLayoutItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(controlsItems))

	for index, item := range controlsItems {
		path := path.AtListIndex(index)

		EditorLayoutValue, EditorLayoutValueDiags := NewEditorLayoutValueFromResponse(path, item)
		diags.Append(EditorLayoutValueDiags...)

		listElementValues[index] = EditorLayoutValue
	}

	list, listDiags := types.ListValue(EditorLayoutValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
