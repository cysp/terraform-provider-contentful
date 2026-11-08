package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateFieldTypeAttributes checks the common App Definition and Extension
// field relationships, descending only into recognized Array items. Unrecognized
// types, including ResourceLink, remain opaque. Callers own location rules and
// choose where to apply the Array item type constraint.
func validateFieldTypeAttributes(field types.Object, valuePath path.Path, summary string) diag.Diagnostics {
	var diags diag.Diagnostics
	if field.IsNull() || field.IsUnknown() {
		return diags
	}

	attributes := field.Attributes()

	kind, ok := attributes["type"].(types.String)
	if !ok || kind.IsNull() || kind.IsUnknown() {
		return diags
	}

	items, hasItems := attributes["items"]
	linkType := attributes["link_type"]

	switch kind.ValueString() {
	case "Array":
		diags.Append(validateArrayFieldAttributes(field, valuePath, summary)...)

		if object, ok := items.(types.Object); ok {
			diags.Append(validateFieldTypeAttributes(object, valuePath.AtName("items"), summary)...)
		}

		return diags
	case "Link":
		if linkType.IsNull() || linkType.Equal(types.StringValue("")) {
			diags.AddAttributeError(valuePath.AtName("link_type"), summary, "Field type Link requires a nonempty link_type.")
		}
	case "Symbol", "Text", "RichText", "Integer", "Number", "Date", "Boolean", "Object", "Location":
		if !linkType.IsNull() && !linkType.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("link_type"), summary, fmt.Sprintf("Field type %q cannot have link_type. Omit this attribute.", kind.ValueString()))
		}
	default:
		return diags
	}

	if hasItems && !items.IsNull() && !items.IsUnknown() {
		diags.AddAttributeError(valuePath.AtName("items"), summary, fmt.Sprintf("Field type %q cannot have items. Omit this attribute.", kind.ValueString()))
	}

	return diags
}

func validateArrayFieldAttributes(field types.Object, valuePath path.Path, summary string) diag.Diagnostics {
	var diags diag.Diagnostics

	attributes := field.Attributes()
	items, hasItems := attributes["items"]
	linkType := attributes["link_type"]

	if hasItems {
		if items.IsNull() {
			diags.AddAttributeError(valuePath.AtName("items"), summary, "An Array field type requires an items definition.")
		}

		if !linkType.IsNull() && !linkType.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("link_type"), summary, "An Array field type cannot have link_type. Configure link_type inside items for linked items.")
		}
	}

	return diags
}

func validateArrayFieldItemType(field types.Object, valuePath path.Path, summary string) diag.Diagnostics {
	var diags diag.Diagnostics
	if field.IsNull() || field.IsUnknown() {
		return diags
	}

	attributes := field.Attributes()
	if !attributes["type"].Equal(types.StringValue("Array")) {
		return diags
	}

	items, ok := attributes["items"].(types.Object)
	if !ok || items.IsNull() || items.IsUnknown() {
		return diags
	}

	itemType, ok := items.Attributes()["type"].(types.String)
	if !ok || itemType.IsNull() || itemType.IsUnknown() {
		return diags
	}

	switch itemType.ValueString() {
	case "Text", "RichText", "Integer", "Number", "Boolean", "Date", "Location", "Object", "Array":
		diags.AddAttributeError(valuePath.AtName("items").AtName("type"), summary, "This field type is not supported for Array items. Use Symbol or Link.")
	}

	return diags
}
