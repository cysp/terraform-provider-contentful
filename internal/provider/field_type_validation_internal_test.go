package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldTypeAttributeValidation(t *testing.T) {
	t.Parallel()

	itemTypes := map[string]attr.Type{"type": types.StringType, "link_type": types.StringType}

	item := types.ObjectValueMust(itemTypes, map[string]attr.Value{"type": types.StringValue("Symbol"), "link_type": types.StringNull()})
	linkItem := types.ObjectValueMust(itemTypes, map[string]attr.Value{"type": types.StringValue("Link"), "link_type": types.StringNull()})

	for _, test := range []struct {
		name       string
		kind, link types.String
		items      types.Object
		paths      []string
	}{
		{"scalar", types.StringValue("Symbol"), types.StringNull(), types.ObjectNull(itemTypes), nil},
		{"scalar link", types.StringValue("Number"), types.StringValue("Entry"), types.ObjectNull(itemTypes), []string{"field.link_type"}},
		{"scalar items", types.StringValue("Symbol"), types.StringNull(), item, []string{"field.items"}},
		{"Link missing target", types.StringValue("Link"), types.StringNull(), types.ObjectNull(itemTypes), []string{"field.link_type"}},
		{"Link empty target", types.StringValue("Link"), types.StringValue(""), types.ObjectNull(itemTypes), []string{"field.link_type"}},
		{"Link unknown target", types.StringValue("Link"), types.StringUnknown(), types.ObjectNull(itemTypes), nil},
		{"Link future target", types.StringValue("Link"), types.StringValue("Future"), types.ObjectNull(itemTypes), nil},
		{"Link items", types.StringValue("Link"), types.StringValue("Entry"), item, []string{"field.items"}},
		{"Array missing items", types.StringValue("Array"), types.StringNull(), types.ObjectNull(itemTypes), []string{"field.items"}},
		{"Array unknown items", types.StringValue("Array"), types.StringNull(), types.ObjectUnknown(itemTypes), nil},
		{"Array incomplete Link items", types.StringValue("Array"), types.StringNull(), linkItem, []string{"field.items.link_type"}},
		{"unknown type with incomplete Link items", types.StringUnknown(), types.StringNull(), linkItem, nil},
		{"future type with incomplete Link items", types.StringValue("Future"), types.StringNull(), linkItem, nil},
		{"ResourceLink with incomplete Link items", types.StringValue("ResourceLink"), types.StringNull(), linkItem, nil},
		{"scalar with incomplete Link items", types.StringValue("Symbol"), types.StringNull(), linkItem, []string{"field.items"}},
		{"Array items", types.StringValue("Array"), types.StringNull(), item, nil},
		{"Array outer target", types.StringValue("Array"), types.StringValue("Entry"), item, []string{"field.link_type"}},
		{"unknown type", types.StringUnknown(), types.StringValue("Entry"), item, nil},
		{"absent type", types.StringNull(), types.StringValue("Entry"), item, nil},
		{"future type", types.StringValue("Future"), types.StringValue("Entry"), item, nil},
		{"ResourceLink is opaque", types.StringValue("ResourceLink"), types.StringNull(), item, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			field := types.ObjectValueMust(map[string]attr.Type{"type": types.StringType, "link_type": types.StringType, "items": types.ObjectType{AttrTypes: itemTypes}}, map[string]attr.Value{"type": test.kind, "link_type": test.link, "items": test.items})

			diags := validateFieldTypeAttributes(field, path.Root("field"), "Invalid field")
			if len(test.paths) == 0 {
				require.Empty(t, diags)

				return
			}

			require.Equal(t, test.paths, mutationDiagnosticPaths(t, diags))

			for _, diagnostic := range diags {
				assert.Equal(t, "Invalid field", diagnostic.Summary())
			}
		})
	}

	for _, field := range []types.Object{types.ObjectNull(itemTypes), types.ObjectUnknown(itemTypes)} {
		require.Empty(t, validateFieldTypeAttributes(field, path.Root("field"), "Invalid field"))
	}
}
