package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// func NewEditorInterfaceEditorLayoutItemValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (EditorInterfaceEditorLayoutItemGroupValue, diag.Diagnostics) {
// 	diags := diag.Diagnostics{}

// 	switch item.Type {
// 	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
// 		groupItem, groupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
// 		if !groupItemOk {
// 			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")

// 			break
// 		}

// 		groupItemItems, groupItemItemsDiags := NewEditorInterfaceEditorLayoutItemGroupItemValueListFromResponse(ctx, path.AtName("items"), groupItem.Items)
// 		diags.Append(groupItemItemsDiags...)

// 		return EditorInterfaceEditorLayoutItemGroupValue{
// 			GroupID: types.StringValue(groupItem.GroupId),
// 			Name:    types.StringValue(groupItem.Name),
// 			Items:   groupItemItems,
// 			state:   attr.ValueStateKnown,
// 		}, diags

// 	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
// 		diags.AddAttributeError(path, "Failed to read editor layout item", "Expected group item")
// 	}

// 	return NewEditorInterfaceEditorLayoutItemGroupValueNull(), diags
// }

// func (v *EditorInterfaceEditorLayoutItemValue) ToEditorInterfaceFieldsEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
// 	diags := diag.Diagnostics{}

// 	item := cm.EditorInterfaceEditorLayoutGroupItem{
// 		GroupId: v.GroupID.ValueString(),
// 		Name:    v.Name.ValueString(),
// 	}

// 	if !v.Items.IsNull() && !v.Items.IsUnknown() {
// 		itemItemsValues := v.Items.Elements()

// 		itemItems := make([]cm.EditorInterfaceEditorLayoutItem, len(itemItemsValues))

// 		for index, itemItem := range itemItemsValues {
// 			itemItemObject, itemItemObjectDiags := itemItem.ToEditorInterfaceEditorLayoutItem(ctx, path.AtListIndex(index))
// 			diags.Append(itemItemObjectDiags...)

// 			itemItems[index] = itemItemObject
// 		}

// 		item.Items = itemItems
// 	}

// 	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(item), diags
// }

func (v *EditorInterfaceEditorLayoutItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// if !v.Field.IsUnknown() && !v.Field.IsNull() {
	// 	fieldItem, fieldItemDiags := v.Field.ToEditorInterfaceEditorLayoutFieldItem(ctx, path.AtName("field"))
	// 	diags.Append(fieldItemDiags...)

	// 	return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
	// }

	if !v.Group.IsUnknown() && !v.Group.IsNull() {
		groupItem, groupItemDiags := v.Group.ToEditorInterfaceEditorLayoutItem(ctx, path.AtName("group"))
		diags.Append(groupItemDiags...)

		return groupItem, diags
	}

	return cm.EditorInterfaceEditorLayoutItem{}, diags
}
