package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutItemGroupItemGroupValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutGroupItem) (EditorInterfaceEditorLayoutItemGroupItemGroupValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	items, itemsDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueListFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsDiags...)

	return EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		GroupID: types.StringValue(item.GroupId),
		Name:    types.StringValue(item.Name),
		Items:   items,
		state:   attr.ValueStateKnown,
	}, diags
}

func (v *EditorInterfaceEditorLayoutItemGroupItemGroupValue) ToEditorInterfaceEditorLayoutGroupItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutGroupItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	groupItem := cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: v.GroupID.ValueString(),
		Name:    v.Name.ValueString(),
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		itemItemsValues := v.Items.Elements()

		itemItems := make([]cm.EditorInterfaceEditorLayoutItem, len(itemItemsValues))

		for index, itemItem := range itemItemsValues {
			itemItemObject, itemItemObjectDiags := itemItem.ToEditorInterfaceEditorLayoutItem(ctx, path.AtListIndex(index))
			diags.Append(itemItemObjectDiags...)

			itemItems[index] = itemItemObject
		}

		groupItem.Items = itemItems
	}

	return groupItem, diags
}
