package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func NewEditorInterfaceEditorLayoutListValueNull(ctx context.Context) TypedList[EditorInterfaceEditorLayoutItemValue] {
	return NewTypedListNull[EditorInterfaceEditorLayoutItemValue](ctx)
}

func NewEditorInterfaceEditorLayoutListValueFromResponse(ctx context.Context, path path.Path, editorLayoutItems []cm.EditorInterfaceEditorLayoutItem) (TypedList[EditorInterfaceEditorLayoutItemValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]EditorInterfaceEditorLayoutItemValue, len(editorLayoutItems))

	for index, item := range editorLayoutItems {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = EditorInterfaceEditorLayoutItemValue{
			Group: editorLayoutValue,
			state: attr.ValueStateKnown,
		}
	}

	list, listDiags := NewTypedList(ctx, listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
