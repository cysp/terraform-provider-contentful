package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func NewEditorInterfaceEditorLayoutListValueFromResponse(ctx context.Context, path path.Path, editorLayoutItems []cm.EditorInterfaceEditorLayoutItem) (TypedList[TypedObject[EditorInterfaceEditorLayoutItemValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[EditorInterfaceEditorLayoutItemValue], len(editorLayoutItems))

	for index, item := range editorLayoutItems {
		path := path.AtListIndex(index)

		editorLayoutValue, editorLayoutValueDiags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(ctx, path, item)
		diags.Append(editorLayoutValueDiags...)

		listElementValues[index] = NewTypedObject(EditorInterfaceEditorLayoutItemValue{
			Group: editorLayoutValue,
		})
	}

	list := NewTypedList(listElementValues)

	return list, diags
}
