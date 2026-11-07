package provider

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type appDefinitionLocationsValidator struct{}

var _ validator.List = appDefinitionLocationsValidator{}

func (appDefinitionLocationsValidator) Description(context.Context) string {
	return "Recognized App Definition locations and field types must use their corresponding nested attributes."
}

func (v appDefinitionLocationsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (appDefinitionLocationsValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	resp.Diagnostics.Append(validateAppDefinitionLocations(req.ConfigValue, req.Path)...)
}

func validateAppDefinitionLocations(locations types.List, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	firstAppConfig := -1

	for index, value := range locations.Elements() {
		location, ok := value.(types.Object)
		if !ok || location.IsNull() || location.IsUnknown() {
			continue
		}

		locationPath := valuePath.AtListIndex(index)
		attributes := location.Attributes()
		diags.Append(validateAppDefinitionLocation(attributes, locationPath)...)

		if attributes["location"].Equal(types.StringValue("app-config")) && attributes["field_types"].IsNull() && attributes["navigation_item"].IsNull() {
			if firstAppConfig >= 0 {
				diags.AddAttributeError(locationPath, "Invalid App Definition location configuration", fmt.Sprintf("This app-config location duplicates locations[%d]. Remove the duplicate location.", firstAppConfig))
			} else {
				firstAppConfig = index
			}
		}

		if fieldTypes, ok := attributes["field_types"].(types.List); ok {
			if attributes["location"].Equal(types.StringValue("entry-field")) {
				diags.Append(validateAppDefinitionRepeatedFields(fieldTypes, locationPath.AtName("field_types"))...)
			}

			for fieldIndex, field := range fieldTypes.Elements() {
				if object, ok := field.(types.Object); ok {
					diags.Append(validateAppDefinitionFieldType(object, locationPath.AtName("field_types").AtListIndex(fieldIndex))...)

					if attributes["location"].Equal(types.StringValue("entry-field")) {
						diags.Append(validateAppDefinitionArrayItem(object, locationPath.AtName("field_types").AtListIndex(fieldIndex))...)
					}
				}
			}
		}

		diags.Append(validateAppDefinitionNavigation(attributes, locationPath)...)
	}

	return diags
}

// Only the two complete shapes with isolated CMA uniqueness evidence are checked.
// Null/unknown children cannot accidentally compare equal as field identities.
func validateAppDefinitionRepeatedFields(fields types.List, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	first := map[string]int{}

	for index, value := range fields.Elements() {
		field, ok := value.(types.Object)
		if !ok || field.IsNull() || field.IsUnknown() {
			continue
		}

		attributes := field.Attributes()
		if !attributes["link_type"].IsNull() {
			continue
		}

		kind := ""
		if attributes["type"].Equal(types.StringValue("Symbol")) && attributes["items"].IsNull() {
			kind = "Symbol"
		} else if attributes["type"].Equal(types.StringValue("Array")) {
			items, ok := attributes["items"].(types.Object)
			if ok && !items.IsNull() && !items.IsUnknown() && items.Attributes()["type"].Equal(types.StringValue("Symbol")) && items.Attributes()["link_type"].IsNull() {
				kind = "Array"
			}
		}

		if kind == "" {
			continue
		}

		if previous, exists := first[kind]; exists {
			diags.AddAttributeError(valuePath.AtListIndex(index), "Invalid App Definition location configuration", fmt.Sprintf("This field definition duplicates %s. Remove the duplicate definition.", valuePath.AtListIndex(previous)))
		} else {
			first[kind] = index
		}
	}

	return diags
}

func validateAppDefinitionNavigation(attributes map[string]attr.Value, locationPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if navigation, ok := attributes["navigation_item"].(types.Object); ok && !navigation.IsNull() && !navigation.IsUnknown() {
		if attributes["location"].Equal(types.StringValue("page")) {
			navPath, ok := navigation.Attributes()["path"].(types.String)
			if ok && !navPath.IsNull() && !navPath.IsUnknown() && navPath.ValueString() != "" && !strings.HasPrefix(navPath.ValueString(), "/") {
				diags.AddAttributeError(locationPath.AtName("navigation_item").AtName("path"), "Invalid App Definition location configuration", "Navigation item path must begin with /, for example /dashboard.")
			}
		}

		for _, name := range []string{"name", "path"} {
			if navigation.Attributes()[name].Equal(types.StringValue("")) {
				diags.AddAttributeError(locationPath.AtName("navigation_item").AtName(name), "Invalid App Definition location configuration", "Navigation item "+name+" must not be empty.")
			}
		}
	}

	return diags
}

func validateAppDefinitionArrayItem(field types.Object, valuePath path.Path) diag.Diagnostics {
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
		diags.AddAttributeError(valuePath.AtName("items").AtName("type"), "Invalid App Definition location configuration", "This field type is not supported for Array items. Use Symbol or Link.")
	}

	return diags
}

func validateAppDefinitionLocation(attributes map[string]attr.Value, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	location, ok := attributes["location"].(types.String)
	if !ok || location.IsNull() || location.IsUnknown() {
		return diags
	}

	switch location.ValueString() {
	case "app-config", "entry-sidebar", "entry-editor", "entry-field", "dialog", "page", "home", "experience-sidebar", "component-sidebar", "experience-toolbar":
		fieldTypes := attributes["field_types"]
		if location.ValueString() == "entry-field" {
			if values, ok := fieldTypes.(types.List); ok && !values.IsUnknown() && len(values.Elements()) == 0 {
				diags.AddAttributeError(valuePath.AtName("field_types"), "Invalid App Definition location configuration", "An entry-field location requires at least one field_types definition.")
			}
		} else if !fieldTypes.IsNull() && !fieldTypes.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("field_types"), "Invalid App Definition location configuration", fmt.Sprintf("field_types cannot be configured for location %q. Omit this attribute.", location.ValueString()))
		}

		navigation := attributes["navigation_item"]
		if location.ValueString() != "page" && !navigation.IsNull() && !navigation.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("navigation_item"), "Invalid App Definition location configuration", fmt.Sprintf("navigation_item cannot be configured for location %q. Omit this attribute.", location.ValueString()))
		}
	}

	return diags
}

func validateAppDefinitionFieldType(field types.Object, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if field.IsNull() || field.IsUnknown() {
		return diags
	}

	attributes := field.Attributes()

	if object, ok := attributes["items"].(types.Object); ok {
		diags.Append(validateAppDefinitionFieldType(object, valuePath.AtName("items"))...)
	}

	diags.Append(validateAppDefinitionFieldAttributes(attributes, valuePath)...)

	return diags
}

func validateAppDefinitionFieldAttributes(attributes map[string]attr.Value, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	fieldType, ok := attributes["type"].(types.String)
	if !ok || fieldType.IsNull() || fieldType.IsUnknown() {
		return diags
	}

	items, hasItems := attributes["items"]
	linkType := attributes["link_type"]

	switch fieldType.ValueString() {
	case "Array":
		// Item type names are API-validated; only outer fields require items.
		if hasItems {
			if items.IsNull() {
				diags.AddAttributeError(valuePath.AtName("items"), "Invalid App Definition location configuration", "An Array field type requires an items definition.")
			}

			if !linkType.IsNull() && !linkType.IsUnknown() {
				diags.AddAttributeError(valuePath.AtName("link_type"), "Invalid App Definition location configuration", "An Array field type cannot have link_type. Configure link_type inside items for linked items.")
			}
		}

		return diags
	case "Link", "ResourceLink":
		if fieldType.ValueString() == "Link" && linkType.Equal(types.StringValue("")) {
			diags.AddAttributeError(valuePath.AtName("link_type"), "Invalid App Definition location configuration", "Field type Link requires a nonempty link_type.")
		}

		if linkType.IsNull() {
			diags.AddAttributeError(valuePath.AtName("link_type"), "Invalid App Definition location configuration", fmt.Sprintf("Field type %q requires link_type.", fieldType.ValueString()))
		}

	case "Symbol", "Text", "RichText", "Integer", "Number", "Date", "Boolean", "Object", "Location":
		if !linkType.IsNull() && !linkType.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("link_type"), "Invalid App Definition location configuration", fmt.Sprintf("Field type %q cannot have link_type. Omit this attribute.", fieldType.ValueString()))
		}

	default:
		return diags
	}

	if hasItems && !items.IsNull() && !items.IsUnknown() {
		diags.AddAttributeError(valuePath.AtName("items"), "Invalid App Definition location configuration", fmt.Sprintf("Field type %q cannot have items. Omit this attribute.", fieldType.ValueString()))
	}

	return diags
}

func validateAppDefinitionPlannedLocations(ctx context.Context, plan tfsdk.Plan) diag.Diagnostics {
	valuePath := path.Root("locations")

	var locations types.List

	diags := plan.GetAttribute(ctx, valuePath, &locations)
	if diags.HasError() {
		return diags
	}

	if locations.IsNull() {
		diags.AddAttributeError(valuePath, "Invalid App Definition location configuration", "locations must be configured. Use an empty list for an app with no locations.")

		return diags
	}

	diags.Append(requireKnownAppDefinitionLocationValue(locations, valuePath)...)
	diags.Append(validateAppDefinitionLocations(locations, valuePath)...)

	return diags
}

// Reject unknown descendants before decoding native slices and pointers so
// diagnostics retain the failing attribute paths.
func requireKnownAppDefinitionLocationValue(value attr.Value, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unknown App Definition location value", "This value must be known before sending the App Definition to Contentful.")

		return diags
	}

	if value.IsNull() {
		return diags
	}

	switch value := value.(type) {
	case types.List:
		for index, element := range value.Elements() {
			elementPath := valuePath.AtListIndex(index)
			if element.IsNull() {
				diags.AddAttributeError(elementPath, "Invalid App Definition location configuration", "List elements must not be null.")
			} else {
				diags.Append(requireKnownAppDefinitionLocationValue(element, elementPath)...)
			}
		}
	case types.Object:
		attributes := value.Attributes()
		for _, name := range slices.Sorted(maps.Keys(attributes)) {
			diags.Append(requireKnownAppDefinitionLocationValue(attributes[name], valuePath.AtName(name))...)
		}
	}

	return diags
}
