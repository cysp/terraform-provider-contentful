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

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(_ context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a group where this layout position requires a field. Terraform state retains a known null value; a later request conversion will reject it until configured.")

		return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a field layout item without its field payload. Terraform state retains a known null value; a later request conversion will reject it until configured.")

			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
			FieldID: types.StringValue(itemFieldItem.FieldId),
		}), diags
	default:
		diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned an unknown editor layout item type. Terraform state retains a known null value; a later request conversion will reject it until configured.")

		return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](), diags
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ToEditorInterfaceEditorLayoutItem(path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	fieldPath := path.AtName("field")
	field, diags := editorInterfaceRequiredObject(v.Field, fieldPath)

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	item, itemDiags := field.ToEditorInterfaceEditorLayoutItem(fieldPath)
	diags.Append(itemDiags...)

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	return item, diags
}
