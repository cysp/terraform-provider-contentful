//nolint:testpackage // Resource identity declarations are intentionally tested through their package-private seam.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceIdentityPathsPreserveTupleOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		[]path.Path{path.Root("space_id"), path.Root("environment_id"), path.Root("entry_id")},
		resourceIdentityPaths([]string{"space_id", "environment_id", "entry_id"}),
	)
}

func TestResourceIdentitySchemaRequiresEveryTupleAttributeForImport(t *testing.T) {
	t.Parallel()

	schema := resourceIdentitySchema([]string{"space_id", "environment_id", "entry_id"})
	require.Len(t, schema.Attributes, 3)

	for _, attributeName := range []string{"space_id", "environment_id", "entry_id"} {
		attribute, ok := schema.Attributes[attributeName].(identityschema.StringAttribute)
		require.True(t, ok, "identity attribute %q is not a string", attributeName)
		assert.True(t, attribute.RequiredForImport)
	}
}
