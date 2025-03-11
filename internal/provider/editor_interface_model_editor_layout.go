package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorInterfaceEditorLayoutValue{}.Type(ctx))
}

func NewEditorInterfaceEditorLayoutValueKnown() EditorInterfaceEditorLayoutValue {
	return EditorInterfaceEditorLayoutValue{
		state: attr.ValueStateKnown,
	}
}

func (v *EditorInterfaceEditorLayoutValue) ToEditorInterfaceFieldsEditorLayoutItem(ctx context.Context, _ path.Path) (cm.EditorInterfaceFieldsEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceFieldsEditorLayoutItem{
		GroupId: v.GroupID.ValueString(),
		Name:    v.Name.ValueString(),
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		var itemItemsStrings []string

		diags.Append(v.Items.ElementsAs(ctx, &itemItemsStrings, false)...)

		itemItems := make([]jx.Raw, len(itemItemsStrings))

		for index, itemItemString := range itemItemsStrings {
			itemItems[index] = jx.Raw(itemItemString)
		}

		item.SetItems(itemItems)
	}

	return item, diags
}

func NewEditorInterfaceEditorLayoutListValueFromResponse(ctx context.Context, path path.Path, editorLayoutItems []cm.EditorInterfaceEditorLayoutItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(editorLayoutItems))

	for index, item := range editorLayoutItems {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutValueFromResponse(path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = editorLayoutValue
	}

	list, listDiags := types.ListValue(EditorInterfaceEditorLayoutValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewEditorInterfaceEditorLayoutValueFromResponse(path path.Path, item cm.EditorInterfaceEditorLayoutItem) (EditorInterfaceEditorLayoutValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutValue{
		GroupID: types.StringValue(item.GroupId),
		Name:    types.StringValue(item.Name),
		state:   attr.ValueStateKnown,
	}

	valueItemsElements := make([]attr.Value, len(item.Items))

	for index, item := range item.Items {
		itemsElement, itemsElementErr := util.JxNormalizeOpaqueBytes(item, util.JxEncodeOpaqueOptions{EscapeStrings: true})

		if itemsElementErr != nil {
			diags.AddAttributeError(path.AtListIndex(index), "Failed to read items element", itemsElementErr.Error())
		}

		valueItemsElements[index] = types.StringValue(string(itemsElement))
	}

	list, listDiags := types.ListValue(types.StringType, valueItemsElements)
	diags.Append(listDiags...)

	value.Items = list

	return value, diags
}
