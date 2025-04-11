package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutListValueNull(ctx context.Context) types.List {
	return types.ListNull(EditorInterfaceEditorLayoutElementValue{}.Type(ctx))
}

func NewEditorInterfaceEditorLayoutElementValueKnown() EditorInterfaceEditorLayoutElementValue {
	return EditorInterfaceEditorLayoutElementValue{
		state: attr.ValueStateKnown,
	}
}

func (v *EditorInterfaceEditorLayoutElementValue) ToEditorInterfaceFieldsEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: v.GroupID.ValueString(),
		Name:    v.Name.ValueString(),
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		var itemItemsValues []EditorInterfaceEditorLayoutElementItemValue

		diags.Append(v.Items.ElementsAs(ctx, &itemItemsValues, false)...)

		itemItems := make([]cm.EditorInterfaceEditorLayoutItem, len(itemItemsValues))

		for index, itemItem := range itemItemsValues {
			itemItemObject, itemItemObjectDiags := itemItem.ToEditorInterfaceEditorLayoutItem(ctx, path.AtListIndex(index))
			diags.Append(itemItemObjectDiags...)

			itemItems[index] = itemItemObject
		}

		item.Items = itemItems
	}

	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(item), diags
}

func (v *EditorInterfaceEditorLayoutElementItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
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

func (v *EditorInterfaceEditorLayoutElementItemFieldValue) ToEditorInterfaceEditorLayoutFieldItem(ctx context.Context, _ path.Path) (cm.EditorInterfaceEditorLayoutFieldItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldItem := cm.EditorInterfaceEditorLayoutFieldItem{
		FieldId: v.FieldID.ValueString(),
	}

	return fieldItem, diags
}
func (v *EditorInterfaceEditorLayoutElementItemGroupValue) ToEditorInterfaceEditorLayoutGroupItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutGroupItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	groupItem := cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: v.GroupID.ValueString(),
		Name:    v.Name.ValueString(),
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		var itemItemsValues []EditorInterfaceEditorLayoutElementItemValue

		diags.Append(v.Items.ElementsAs(ctx, &itemItemsValues, false)...)

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

func NewEditorInterfaceEditorLayoutListValueFromResponse(ctx context.Context, path path.Path, editorLayoutItems []cm.EditorInterfaceEditorLayoutItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(editorLayoutItems))

	for index, item := range editorLayoutItems {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutElementValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = editorLayoutValue
	}

	list, listDiags := types.ListValue(EditorInterfaceEditorLayoutElementValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewEditorInterfaceEditorLayoutElementValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (EditorInterfaceEditorLayoutElementValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		groupItem, groupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
		if !groupItemOk {
			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")
			return EditorInterfaceEditorLayoutElementValue{}, diags
		}

		groupItemItems, groupItemItemsDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(ctx, path.AtName("items"), groupItem.Items)
		diags.Append(groupItemItemsDiags...)

		return EditorInterfaceEditorLayoutElementValue{
			GroupID: types.StringValue(groupItem.GroupId),
			Name:    types.StringValue(groupItem.Name),
			Items:   groupItemItems,
			state:   attr.ValueStateKnown,
		}, diags

	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
	}

	diags.AddAttributeError(path, "Failed to read editor layout item", "Expected group item")
	return EditorInterfaceEditorLayoutElementValue{}, diags
}

func NewEditorInterfaceEditorLayoutElementItemValueListFromResponse(ctx context.Context, path path.Path, items []cm.EditorInterfaceEditorLayoutItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutElementItemValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = editorLayoutValue
	}

	list, listDiags := types.ListValue(EditorInterfaceEditorLayoutElementItemValue{}.Type(ctx), listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewEditorInterfaceEditorLayoutElementItemValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutItem) (EditorInterfaceEditorLayoutElementItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.EditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem:
		itemGroupItem, itemGroupItemOk := item.GetEditorInterfaceEditorLayoutGroupItem()
		if !itemGroupItemOk {
			diags.AddAttributeError(path, "Failed to read group item", "Expected group item")
			return EditorInterfaceEditorLayoutElementItemValue{}, diags
		}

		groupValue, groupValueDiags := NewEditorInterfaceEditorLayoutElementItemGroupValueFromResponse(ctx, path, itemGroupItem)
		diags.Append(groupValueDiags...)

		return EditorInterfaceEditorLayoutElementItemValue{
			Group: groupValue,
			state: attr.ValueStateKnown,
		}, diags
	case cm.EditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem:
		itemFieldItem, itemFieldItemOk := item.GetEditorInterfaceEditorLayoutFieldItem()
		if !itemFieldItemOk {
			diags.AddAttributeError(path, "Failed to read field item", "Expected field item")
			return EditorInterfaceEditorLayoutElementItemValue{}, diags
		}

		fieldValue, fieldValueDiags := NewEditorInterfaceEditorLayoutElementItemFieldValueFromResponse(ctx, path, itemFieldItem)
		diags.Append(fieldValueDiags...)

		return EditorInterfaceEditorLayoutElementItemValue{
			Field: fieldValue,
			state: attr.ValueStateKnown,
		}, diags
	default:
		return EditorInterfaceEditorLayoutElementItemValue{}, diags
	}
}

func NewEditorInterfaceEditorLayoutElementItemGroupValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutGroupItem) (EditorInterfaceEditorLayoutElementItemGroupValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	items, itemsDiags := NewEditorInterfaceEditorLayoutElementItemValueListFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsDiags...)

	return EditorInterfaceEditorLayoutElementItemGroupValue{
		GroupID: types.StringValue(item.GroupId),
		Name:    types.StringValue(item.Name),
		Items:   items,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewEditorInterfaceEditorLayoutElementItemFieldValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutFieldItem) (EditorInterfaceEditorLayoutElementItemFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	return EditorInterfaceEditorLayoutElementItemFieldValue{
		FieldID: types.StringValue(item.FieldId),
		state:   attr.ValueStateKnown,
	}, diags
}
