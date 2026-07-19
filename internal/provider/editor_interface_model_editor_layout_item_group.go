package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		groupItem, groupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
		if !groupItemOk {
			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")

			break
		}

		groupItemItems, groupItemItemsDiags := NewEditorInterfaceEditorLayoutItemGroupItemValueListFromResponse(ctx, path.AtName("items"), groupItem.Items)
		diags.Append(groupItemItemsDiags...)

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupValue{
			GroupID: types.StringValue(groupItem.GroupId),
			Name:    types.StringValue(groupItem.Name),
			Items:   groupItemItems,
		}), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		diags.AddAttributeError(path, "Failed to read editor layout item", "Expected group item")
	}

	return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](), diags
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, valuePath path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: v.GroupID.ValueString(),
		Name:    v.Name.ValueString(),
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		itemItems, itemDiags := ConvertKnownObjectListElements(
			ctx,
			valuePath.AtName("items"),
			v.Items.Elements(),
			func(ctx context.Context, itemPath path.Path, value EditorInterfaceEditorLayoutItemGroupItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				return value.ToEditorInterfaceEditorLayoutItem(ctx, itemPath)
			},
		)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			item.Items = itemItems
		}
	}

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(item), diags
}
