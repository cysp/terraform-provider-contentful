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

		if diags.HasError() {
			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupValue{
			GroupID: types.StringValue(groupItem.GroupId),
			Name:    types.StringValue(groupItem.Name),
			Items:   groupItemItems,
		}), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		diags.AddAttributeError(path, "Failed to read editor layout item", "Expected group item")
	default:
		diags.AddAttributeError(path, "Failed to read editor layout item", "Contentful returned an unknown editor layout item type.")
	}

	return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](), diags
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, valuePath path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	item, diags := toEditorInterfaceEditorLayoutGroupItem(
		ctx,
		valuePath,
		v.GroupID,
		v.Name,
		v.Items,
		func(ctx context.Context, itemPath path.Path, value EditorInterfaceEditorLayoutItemGroupItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
			return value.ToEditorInterfaceEditorLayoutItem(ctx, itemPath)
		},
	)
	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(item), diags
}
