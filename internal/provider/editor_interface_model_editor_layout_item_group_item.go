package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func NewEditorInterfaceEditorLayoutItemGroupItemValueListFromResponse(ctx context.Context, path path.Path, items []cm.EditorInterfaceEditorLayoutItem) (TypedList[EditorInterfaceEditorLayoutItemGroupItemValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]EditorInterfaceEditorLayoutItemGroupItemValue, len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutItemValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = editorLayoutValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceEditorLayoutItemValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (EditorInterfaceEditorLayoutItemGroupItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		itemGroupItem, itemGroupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
		if !itemGroupItemOk {
			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")

			return EditorInterfaceEditorLayoutItemGroupItemValue{}, diags
		}

		groupValue, groupValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupValueFromResponse(ctx, path, itemGroupItem)
		diags.Append(groupValueDiags...)

		return EditorInterfaceEditorLayoutItemGroupItemValue{
			Group: groupValue,
			state: attr.ValueStateKnown,
		}, diags
	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeError(path, "Failed to read field item", "Expected field item")

			return EditorInterfaceEditorLayoutItemGroupItemValue{}, diags
		}

		fieldValue, fieldValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemFieldValueFromResponse(ctx, path, itemFieldItem)
		diags.Append(fieldValueDiags...)

		return EditorInterfaceEditorLayoutItemGroupItemValue{
			Field: fieldValue,
			state: attr.ValueStateKnown,
		}, diags
	default:
		return EditorInterfaceEditorLayoutItemGroupItemValue{}, diags
	}
}

func (v *EditorInterfaceEditorLayoutItemGroupItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !v.Field.IsUnknown() && !v.Field.IsNull() {
		fieldItem, fieldItemDiags := v.Field.ToEditorInterfaceEditorLayoutFieldItem(ctx, path.AtName("field"))
		diags.Append(fieldItemDiags...)

		return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
	}

	if !v.Group.IsUnknown() && !v.Group.IsNull() {
		groupItem, groupItemDiags := v.Group.ToEditorInterfaceEditorLayoutGroupItem(ctx, path.AtName("group"))
		diags.Append(groupItemDiags...)

		return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(groupItem), diags
	}

	return cm.EditorInterfaceEditorLayoutItem{}, diags
}
