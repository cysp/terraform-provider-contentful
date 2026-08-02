package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryRequestDistinguishesTerraformNullFromJSONNull(t *testing.T) {
	t.Parallel()

	model := EntryModel{
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"terraform_null": jsontypes.NewNormalizedNull(),
			"json_null":      NewNormalizedJSONTypesNormalizedValue([]byte("null")),
			"value":          NewNormalizedJSONTypesNormalizedValue([]byte(`{"en-US":"value"}`)),
		}),
		Metadata: NewTypedObjectNull[EntryMetadataValue](),
	}

	request, diags := model.ToEntryRequest(t.Context())
	require.False(t, diags.HasError(), diags.Errors())

	fields, ok := request.Fields.Get()
	require.True(t, ok)
	assert.NotContains(t, fields, "terraform_null")
	assert.JSONEq(t, "null", string(fields["json_null"]))
	assert.JSONEq(t, `{"en-US":"value"}`, string(fields["value"]))
}

func TestEntryRequestUnknownFieldFailsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	model := EntryModel{
		Fields: NewTypedMap(map[string]jsontypes.Normalized{
			"known":   NewNormalizedJSONTypesNormalizedValue([]byte(`"value"`)),
			"unknown": jsontypes.NewNormalizedUnknown(),
		}),
		Metadata: NewTypedObjectNull[EntryMetadataValue](),
	}

	request, diags := model.ToEntryRequest(t.Context())
	assert.False(t, request.Fields.IsSet())
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`fields["unknown"]`}, diagnosticPaths(t, diags))
}
