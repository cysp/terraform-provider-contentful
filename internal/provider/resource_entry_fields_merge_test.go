//nolint:testpackage
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
)

func TestMergeMissingEntryFieldsInitializesNullTarget(t *testing.T) {
	t.Parallel()

	target := NewTypedMapNull[TypedMap[jsontypes.Normalized]]()
	fallback := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
		"name": NewTypedMap(map[string]jsontypes.Normalized{
			"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"name"`)),
		}),
	})

	mergeMissingEntryFields(&target, fallback)

	assert.False(t, target.IsNull())
	assert.Equal(t, `"name"`, target.Elements()["name"].Elements()["en-AU"].ValueString())
}

func TestMergeMissingEntryFieldsPreservesExistingValues(t *testing.T) {
	t.Parallel()

	target := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
		"name": NewTypedMap(map[string]jsontypes.Normalized{
			"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"server"`)),
		}),
	})
	fallback := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
		"name": NewTypedMap(map[string]jsontypes.Normalized{
			"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"plan"`)),
		}),
		"blurb": NewTypedMap(map[string]jsontypes.Normalized{
			"en-AU": NewNormalizedJSONTypesNormalizedValue([]byte(`"added"`)),
		}),
	})

	mergeMissingEntryFields(&target, fallback)

	assert.Equal(t, `"server"`, target.Elements()["name"].Elements()["en-AU"].ValueString())
	assert.Equal(t, `"added"`, target.Elements()["blurb"].Elements()["en-AU"].ValueString())
}
