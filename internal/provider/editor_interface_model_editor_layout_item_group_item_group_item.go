package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueListFromResponse(ctx context.Context, path path.Path, items []cm.EditorInterfaceEditorLayoutItem) (TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue], len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
			Field: editorLayoutValue,
		})
	}

	if diags.HasError() {
		return NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]](), diags
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(_ context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		diags.AddAttributeError(path, "Failed to read editor interface editor layout", "Unexpected group item")

		return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeError(path, "Failed to read field item", "Expected field item")

			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
			FieldID: types.StringValue(itemFieldItem.FieldId),
		}), diags
	default:
		diags.AddAttributeError(path, "Failed to read editor interface editor layout", "Contentful returned an unknown editor layout item type.")

		return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldPath := path.AtName("field")
	field, fieldDiags := KnownObjectValue(v.Field, fieldPath)
	diags.Append(fieldDiags...)

	if !fieldDiags.HasError() {
		fieldItem, fieldItemDiags := field.ToEditorInterfaceEditorLayoutItem(ctx, fieldPath)
		diags.Append(fieldItemDiags...)

		return fieldItem, diags
	}

	// if !v.Group.IsUnknown() && !v.Group.IsNull() {
	// 	groupItem, groupItemDiags := v.Group.Value().ToEditorInterfaceEditorLayoutGroupItem(ctx, path.AtName("group"))
	// 	diags.Append(groupItemDiags...)

	// 	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(groupItem), diags
	// }

	return cm.EditorInterfaceEditorLayoutItem{}, diags
}
