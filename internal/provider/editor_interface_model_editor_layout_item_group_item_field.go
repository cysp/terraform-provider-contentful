package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceEditorLayoutItemGroupItemFieldValueFromResponse(_ context.Context, _ path.Path, item cm.EditorInterfaceEditorLayoutFieldItem) (TypedObject[EditorInterfaceEditorLayoutItemGroupItemFieldValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	return NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemFieldValue{
		FieldID: types.StringValue(item.FieldId),
	}), diags
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) ToEditorInterfaceEditorLayoutItem(_ context.Context, _ path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldItem := cm.EditorInterfaceEditorLayoutFieldItem{
		FieldId: v.FieldID.ValueString(),
	}

	return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) ToEditorInterfaceEditorLayoutFieldItem(_ context.Context, _ path.Path) (cm.EditorInterfaceEditorLayoutFieldItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldItem := cm.EditorInterfaceEditorLayoutFieldItem{
		FieldId: v.FieldID.ValueString(),
	}

	return fieldItem, diags
}
