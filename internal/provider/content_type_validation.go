package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func contentTypeJSONObjectValidator() jsonObjectValidator {
	return jsonObjectValidator{summary: "Invalid Content Type JSON object", detail: "The value must encode a JSON object. Omit an optional value instead of encoding JSON null."}
}

func validateContentTypeFieldIDs(fields TypedList[TypedObject[ContentTypeFieldValue]], valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	seen := map[string]bool{}

	for index, element := range fields.Elements() {
		field, ok := element.GetValue()
		if !ok || field.ID.IsNull() || field.ID.IsUnknown() {
			continue
		}

		id := field.ID.ValueString()
		if seen[id] {
			diags.AddAttributeError(valuePath.AtListIndex(index).AtName("id"), "Duplicate Content Type field ID", "Field IDs must be unique within the Content Type.")
		}

		seen[id] = true
	}

	return diags
}
