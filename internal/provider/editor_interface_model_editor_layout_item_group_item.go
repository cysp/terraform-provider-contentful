package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func NewEditorInterfaceEditorLayoutItemGroupItemValueListFromResponse(ctx context.Context, path path.Path, items []cm.EditorInterfaceEditorLayoutItem) (TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue], len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutItemValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = editorLayoutValue
	}

	if diags.HasError() {
		return NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]](), diags
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceEditorLayoutItemValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		itemGroupItem, itemGroupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
		if !itemGroupItemOk {
			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")

			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
		}

		groupValue, groupValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupValueFromResponse(ctx, path, itemGroupItem)
		diags.Append(groupValueDiags...)

		if diags.HasError() {
			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
			Group: groupValue,
		}), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeError(path, "Failed to read field item", "Expected field item")

			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
		}

		fieldValue, fieldValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemFieldValueFromResponse(ctx, path, itemFieldItem)
		diags.Append(fieldValueDiags...)

		if diags.HasError() {
			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Field: fieldValue,
			Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
		}), diags

	default:
		diags.AddAttributeError(path, "Failed to read editor layout item", "Contentful returned an unknown editor layout item type.")

		return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, valuePath path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	fieldPath := valuePath.AtName("field")
	groupPath := valuePath.AtName("group")

	return ConvertExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative[cm.EditorInterfaceEditorLayoutItem]{
			Name: "field", Path: fieldPath, Value: v.Field,
			Convert: func() (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				field, _ := v.Field.GetValue()
				fieldItem, diags := field.ToEditorInterfaceEditorLayoutFieldItem(ctx, fieldPath)

				return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
			},
		},
		KnownUnionAlternative[cm.EditorInterfaceEditorLayoutItem]{
			Name: "group", Path: groupPath, Value: v.Group,
			Convert: func() (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				group, _ := v.Group.GetValue()
				groupItem, diags := group.ToEditorInterfaceEditorLayoutGroupItem(ctx, groupPath)

				return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(groupItem), diags
			},
		},
	)
}
