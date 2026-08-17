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
			diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a group layout item without its group payload. Terraform state retains a known null value; a later request conversion will reject it until configured.")

			return NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](), diags
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
		diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned a field where this layout position requires a group. Terraform state retains a known null value; a later request conversion will reject it until configured.")
	default:
		diags.AddAttributeWarning(path, "Unsupported editor layout item", "Contentful returned an unknown editor layout item type. Terraform state retains a known null value; a later request conversion will reject it until configured.")
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
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceEditorLayoutItemGroupItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
			return value.ToEditorInterfaceEditorLayoutItem(ctx, valuePath)
		},
	)
	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	return cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(item), diags
}
