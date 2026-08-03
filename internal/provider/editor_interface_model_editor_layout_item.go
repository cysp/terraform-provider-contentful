package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v EditorInterfaceEditorLayoutItemValue) ToEditorInterfaceEditorLayoutItem(ctx context.Context, path path.Path) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
	groupPath := path.AtName("group")
	group, diags := editorInterfaceRequiredObject(v.Group, groupPath)

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	item, itemDiags := group.ToEditorInterfaceEditorLayoutItem(ctx, groupPath)
	diags.Append(itemDiags...)

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutItem{}, diags
	}

	return item, diags
}
