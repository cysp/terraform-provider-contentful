package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EditorInterfaceModel) ToEditorInterfaceData(ctx context.Context) (cm.EditorInterfaceData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.EditorInterfaceData{}

	if model.EditorLayout.IsNull() || model.EditorLayout.IsUnknown() {
		request.EditorLayout.Reset()
	} else {
		items, itemDiags := ConvertKnownObjectListElements(
			ctx,
			path.Root("editor_layout"),
			model.EditorLayout.Elements(),
			func(ctx context.Context, itemPath path.Path, value EditorInterfaceEditorLayoutItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
				return value.ToEditorInterfaceEditorLayoutItem(ctx, itemPath)
			},
		)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			request.EditorLayout.SetTo(items)
		}
	}

	if model.Controls.IsNull() || model.Controls.IsUnknown() {
		request.Controls.Reset()
	} else {
		items, itemDiags := ConvertKnownObjectListElements(
			ctx,
			path.Root("controls"),
			model.Controls.Elements(),
			func(ctx context.Context, itemPath path.Path, value EditorInterfaceControlValue) (cm.EditorInterfaceDataControlsItem, diag.Diagnostics) {
				return value.ToEditorInterfaceDataControlsItem(ctx, itemPath)
			},
		)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			request.Controls.SetTo(items)
		}
	}

	if model.GroupControls.IsNull() || model.GroupControls.IsUnknown() {
		request.GroupControls.Reset()
	} else {
		items, itemDiags := ConvertKnownObjectListElements(
			ctx,
			path.Root("group_controls"),
			model.GroupControls.Elements(),
			func(ctx context.Context, itemPath path.Path, value EditorInterfaceGroupControlValue) (cm.EditorInterfaceDataGroupControlsItem, diag.Diagnostics) {
				return value.ToEditorInterfaceDataGroupControlsItem(ctx, itemPath)
			},
		)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			request.GroupControls.SetTo(items)
		}
	}

	if model.Sidebar.IsNull() || model.Sidebar.IsUnknown() {
		request.Sidebar.Reset()
	} else {
		items, itemDiags := ConvertKnownObjectListElements(
			ctx,
			path.Root("sidebar"),
			model.Sidebar.Elements(),
			func(ctx context.Context, itemPath path.Path, value EditorInterfaceSidebarValue) (cm.EditorInterfaceDataSidebarItem, diag.Diagnostics) {
				return value.ToEditorInterfaceDataSidebarItem(ctx, itemPath)
			},
		)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			request.Sidebar.SetTo(items)
		}
	}

	if diags.HasError() {
		return cm.EditorInterfaceData{}, diags
	}

	return request, diags
}
