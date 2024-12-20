package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/cysp/terraform-provider-contentful/internal/tf"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorLayoutListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorLayoutValue{}.Type(ctx))
}

func NewEditorLayoutValueKnown() EditorLayoutValue {
	return EditorLayoutValue{
		state: attr.ValueStateKnown,
	}
}

func (model *EditorLayoutValue) ToPutEditorInterfaceReqEditorLayoutItem(ctx context.Context, _ path.Path) (contentfulManagement.PutEditorInterfaceReqEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := contentfulManagement.PutEditorInterfaceReqEditorLayoutItem{
		GroupId: model.GroupId.ValueString(),
		Name:    model.Name.ValueString(),
	}

	if tf.IsKnownAndPresent(model.Items) {
		var itemItemsStrings []string

		diags.Append(model.Items.ElementsAs(ctx, &itemItemsStrings, false)...)

		itemItems := make([]jx.Raw, len(itemItemsStrings))

		for index, itemItemString := range itemItemsStrings {
			itemItems[index] = jx.Raw(itemItemString)
		}

		item.SetItems(itemItems)
	}

	return item, diags
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

func NewEditorLayoutValueFromResponse(path path.Path, item contentfulManagement.EditorInterfaceEditorLayoutItem) (EditorLayoutValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorLayoutValue{
		GroupId: types.StringValue(item.GroupId),
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
