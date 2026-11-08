package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func editorInterfaceSettingsValidator() jsonObjectValidator {
	return jsonObjectValidator{summary: "Invalid Editor Interface settings", detail: "Widget settings must be a JSON object. Omit settings to leave them unspecified."}
}

// Only compare relationships whose participants are present and known. Omitted
// layout/controls can depend on remote defaults and are not an empty layout.
func validateEditorInterfaceLayout(model EditorInterfaceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	groups := map[string]bool{}
	addGroup := func(groupID types.String, nested bool, valuePath path.Path) {
		if groupID.IsNull() || groupID.IsUnknown() {
			return
		}

		if _, exists := groups[groupID.ValueString()]; exists {
			diags.AddAttributeError(valuePath, "Duplicate Editor Interface group ID", "Group IDs must be unique across the editor layout, including nested groups.")
		} else {
			groups[groupID.ValueString()] = nested
		}
	}

	for index, item := range model.EditorLayout.Elements() {
		layout, ok := item.GetValue()
		if !ok {
			continue
		}

		group, ok := layout.Group.GetValue()
		if !ok {
			continue
		}

		groupPath := path.Root("editor_layout").AtListIndex(index).AtName("group")
		addGroup(group.GroupID, false, groupPath.AtName("group_id"))

		for childIndex, item := range group.Items.Elements() {
			child, ok := item.GetValue()
			if !ok {
				continue
			}

			nested, ok := child.Group.GetValue()
			if !ok {
				continue
			}

			addGroup(nested.GroupID, true, groupPath.AtName("items").AtListIndex(childIndex).AtName("group").AtName("group_id"))
		}
	}
	// A duplicate group ID makes its position ambiguous; report that first.
	if diags.HasError() {
		return diags
	}

	diags.Append(validateEditorInterfaceGroupControls(model.GroupControls, groups)...)

	return diags
}

func validateEditorInterfaceGroupControls(controls TypedList[TypedObject[EditorInterfaceGroupControlValue]], groups map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics

	for index, item := range controls.Elements() {
		control, ok := item.GetValue()
		if !ok {
			continue
		}

		if control.GroupID.IsNull() || control.GroupID.IsUnknown() || control.WidgetNamespace.IsUnknown() || control.WidgetNamespace.IsNull() || control.WidgetNamespace.ValueString() != "builtin" || control.WidgetID.IsNull() || control.WidgetID.IsUnknown() {
			continue
		}

		nested, found := groups[control.GroupID.ValueString()]
		if !found {
			continue
		}

		widget := control.WidgetID.ValueString()
		if nested && widget == "topLevelTab" || !nested && widget == "fieldset" {
			diags.AddAttributeError(path.Root("group_controls").AtListIndex(index).AtName("widget_id"), "Invalid Editor Interface group control", "Use topLevelTab for top-level groups and fieldset for nested groups.")
		}
	}

	return diags
}
