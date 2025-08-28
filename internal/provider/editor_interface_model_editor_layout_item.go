package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v EditorInterfaceEditorLayoutItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// if !v.Field.IsUnknown() && !v.Field.IsNull() {
	// 	fieldItem, fieldItemDiags := v.Field.Value().ToEditorInterfaceEditorLayoutFieldItem(ctx, path.AtName("field"))
	// 	diags.Append(fieldItemDiags...)

	// 	return cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(fieldItem), diags
	// }

	if !v.Group.IsUnknown() && !v.Group.IsNull() {
		groupItem, groupItemDiags := v.Group.Value().ToEditorInterfaceEditorLayoutItem(ctx, path.AtName("group"))
		diags.Append(groupItemDiags...)

		return groupItem, diags
	}

	return cm.EditorInterfaceEditorLayoutItem{}, diags
}
