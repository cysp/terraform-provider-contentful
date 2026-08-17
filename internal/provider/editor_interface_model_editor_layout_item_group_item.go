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
			diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a group layout item without its group payload. Terraform state retains a known empty union; a later request conversion will reject it until configured.")

			return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			}), diags
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
			diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a field layout item without its field payload. Terraform state retains a known empty union; a later request conversion will reject it until configured.")

			return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
				Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
				Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			}), diags
		}

		fieldValue, fieldValueDiags := NewEditorInterfaceEditorLayoutItemGroupItemFieldValueFromResponse(ctx, path, itemFieldItem)
		diags.Append(fieldValueDiags...)

		if diags.HasError() {
			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](), diags
		}

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			Field: fieldValue,
		}), diags

	default:
		diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned an unknown editor layout item type. Terraform state retains a known empty union; a later request conversion will reject it until configured.")

		return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemValue{
			Field: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
			Group: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
		}), diags
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
