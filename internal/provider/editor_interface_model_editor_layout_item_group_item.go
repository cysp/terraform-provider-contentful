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

			return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{}), diags
		}

		groupValue, groupValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupValueFromResponse(ctx, path, itemGroupItem)
		diags.Append(groupValueDiags...)

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Group: groupValue,
		}), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeError(path, "Failed to read field item", "Expected field item")

			return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{}), diags
		}

		fieldValue, fieldValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemFieldValueFromResponse(ctx, path, itemFieldItem)
		diags.Append(fieldValueDiags...)

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Field: fieldValue,
		}), diags

	default:
		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{}), diags
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, valuePath path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	fieldPath := valuePath.AtName("field")
	groupPath := valuePath.AtName("group")

	return convertExactlyOneKnownAlternative(
		valuePath,
		knownUnionAlternative[cm.EditorInterfaceEditorLayoutItem]{
			Name:  "field",
			Path:  fieldPath,
			Value: v.Field,
			Convert: func() (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				field, _ := v.Field.GetValue()
				fieldItem, diags := field.ToEditorInterfaceEditorLayoutFieldItem(fieldPath)

				if diags.HasError() {
					return cm.EditorInterfaceEditorLayoutItem{}, diags
				}

				return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
			},
		},
		knownUnionAlternative[cm.EditorInterfaceEditorLayoutItem]{
			Name:  "group",
			Path:  groupPath,
			Value: v.Group,
			Convert: func() (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				group, _ := v.Group.GetValue()
				groupItem, diags := group.ToEditorInterfaceEditorLayoutGroupItem(ctx, groupPath)

				if diags.HasError() {
					return cm.EditorInterfaceEditorLayoutItem{}, diags
				}

				return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(groupItem), diags
			},
		},
	)
}
