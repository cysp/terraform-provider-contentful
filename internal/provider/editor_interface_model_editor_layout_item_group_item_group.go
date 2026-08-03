package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutItemGroupItemGroupValueFromResponse(ctx context.Context, path path.Path, item cm.EditorInterfaceEditorLayoutGroupItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	items, itemsDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueListFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsDiags...)

	return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		GroupID: types.StringValue(item.GroupId),
		Name:    types.StringValue(item.Name),
		Items:   items,
	}), diags
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) ToEditorInterfaceEditorLayoutGroupItem(ctx context.Context, valuePath path.Path) (cm.EditorInterfaceEditorLayoutGroupItem, diag.Diagnostics) {
	return toEditorInterfaceEditorLayoutGroupItem(
		ctx,
		valuePath,
		v.GroupID,
		v.Name,
		v.Items,
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
			return value.ToEditorInterfaceEditorLayoutItem(ctx, valuePath)
		},
	)
}
