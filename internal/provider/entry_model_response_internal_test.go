package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntryFieldsFromResponsePreservesOmissionAsNull(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(t.Context(), path.Root("fields"), cm.OptEntryFields{})

	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, fields.IsNull())
	assert.False(t, fields.IsUnknown())
	assert.Empty(t, fields.Elements())
}

func TestNewEntryFieldsFromResponsePreservesKnownJSONNull(t *testing.T) {
	t.Parallel()

	fields, diags := NewEntryFieldsFromResponse(t.Context(), path.Root("fields"), cm.NewOptEntryFields(cm.EntryFields{
		"optional": jx.Raw(`null`),
	}))

	require.False(t, diags.HasError(), diags.Errors())
	require.False(t, fields.IsNull())
	require.Contains(t, fields.Elements(), "optional")
	assert.False(t, fields.Elements()["optional"].IsNull())
	assert.Equal(t, `null`, fields.Elements()["optional"].ValueString())
}

func TestMergeEntryFieldsWithFallback(t *testing.T) {
	t.Parallel()

	returnedValue := jsontypes.NewNormalizedValue(`{"en-US":"returned"}`)
	configuredValue := jsontypes.NewNormalizedValue(`{"en-US":"configured"}`)
	missingValue := jsontypes.NewNormalizedValue(`{"en-US":"missing"}`)
	returned := NewTypedMap(map[string]jsontypes.Normalized{"returned": returnedValue})
	configured := NewTypedMap(map[string]jsontypes.Normalized{
		"returned": configuredValue,
		"missing":  missingValue,
	})

	actual := mergeEntryFieldsWithFallback(returned, configured)

	assert.Equal(t, map[string]jsontypes.Normalized{
		"returned": returnedValue,
		"missing":  missingValue,
	}, actual.Elements())
	assert.Equal(t, map[string]jsontypes.Normalized{"returned": returnedValue}, returned.Elements())
	assert.Equal(t, map[string]jsontypes.Normalized{
		"returned": configuredValue,
		"missing":  missingValue,
	}, configured.Elements())
}

func TestMergeEntryFieldsWithFallbackInitializesNullResponse(t *testing.T) {
	t.Parallel()

	value := jsontypes.NewNormalizedValue(`{"en-US":"configured"}`)
	actual := mergeEntryFieldsWithFallback(
		NewTypedMapNull[jsontypes.Normalized](),
		NewTypedMap(map[string]jsontypes.Normalized{"field": value}),
	)

	assert.False(t, actual.IsNull())
	assert.Equal(t, map[string]jsontypes.Normalized{"field": value}, actual.Elements())
}

//nolint:maintidx // One dense table keeps response projection policies comparable.
func TestProjectEntryMutationResponse(t *testing.T) {
	t.Parallel()

	managed := jsontypes.NewNormalizedValue(`{"en-US":"managed"}`)
	external := jsontypes.NewNormalizedValue(`{"en-US":"external"}`)
	changed := jsontypes.NewNormalizedValue(`{"en-US":"changed"}`)
	defaulted := jsontypes.NewNormalizedValue(`{"en-US":"default"}`)
	terraformNull := jsontypes.NewNormalizedNull()
	jsonNull := jsontypes.NewNormalizedValue(`null`)
	emptyArray := jsontypes.NewNormalizedValue(`{"en-US":[]}`)
	semanticPlanValue := jsontypes.NewNormalizedValue(`{"en-US":{"first":"one","second":"two"}}`)
	semanticResponseValue := jsontypes.NewNormalizedValue(`{ "en-US": { "second": "two", "first": "one" } }`)
	remoteMetadata := NewTypedObject[EntryMetadataValue](EntryMetadataValue{
		Concepts: NewTypedListFromStringSlice([]string{"remote"}),
		Tags:     NewTypedListFromStringSlice([]string{}),
	})
	plan := EntryModel{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("article"),
		PublishedVersion:   types.Int64Null(),
		Fields:             NewTypedMap(map[string]jsontypes.Normalized{"managed": managed}),
		Metadata: NewTypedObject[EntryMetadataValue](EntryMetadataValue{
			Concepts: NewTypedListFromStringSlice([]string{}),
			Tags:     NewTypedListFromStringSlice([]string{"managed"}),
		}),
		Timeouts: TimeoutsNull(),
	}

	tests := map[string]struct {
		plan                  EntryModel
		response              EntryModel
		policy                entryResponseFieldPolicy
		hasError              bool
		expectedFields        TypedMap[jsontypes.Normalized]
		keepsResponseMetadata bool
	}{
		"semantically equal JSON": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": semanticPlanValue})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": semanticResponseValue})

				return value
			}(),
			policy:         entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{"managed": semanticPlanValue}),
		},
		"whole fields member omitted": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMapNull[jsontypes.Normalized]()

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMapNull[jsontypes.Normalized](),
		},
		"present partial response missing nonempty field": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "external": external})

				return value
			}(),
			response: plan, policy: entryResponseFieldsCreationDefaults, hasError: true, expectedFields: plan.Fields,
		},
		"present partial response restores all-empty localized array": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "empty": emptyArray})

				return value
			}(),
			response: plan, policy: entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "empty": emptyArray}),
		},
		"additional response field accepted": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsCreationDefaults, expectedFields: plan.Fields,
		},
		"response default for omitted Terraform null accepted on create": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": terraformNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsCreationDefaults,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{
				"managed": managed, "default": terraformNull,
			}),
		},
		"omitted Terraform null accepted by exact response": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": terraformNull})

				return value
			}(),
			response: plan, policy: entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{
				"managed": managed, "optional": terraformNull,
			}),
		},
		"response value for omitted Terraform null rejected by exact response": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": terraformNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": defaulted})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{
				"managed": managed, "optional": defaulted,
			}),
		},
		"missing sent JSON null restored with create default": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": jsonNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsCreationDefaults,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{
				"managed": managed, "optional": jsonNull,
			}),
		},
		"response value for sent JSON null rejected": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": jsonNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "optional": defaulted})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{
				"managed": managed, "optional": defaulted,
			}),
		},
		"additional response field rejected by exact policy": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{"managed": managed, "default": defaulted}),
		},
		"changed managed field": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]jsontypes.Normalized{"managed": changed})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]jsontypes.Normalized{"managed": changed}),
		},
		"metadata contradiction": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Metadata = remoteMetadata

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true, expectedFields: plan.Fields, keepsResponseMetadata: true,
		},
		"metadata order canonicalization": {
			plan: func() EntryModel {
				value := plan
				value.Metadata = NewTypedObject(EntryMetadataValue{
					Concepts: NewTypedListFromStringSlice([]string{"first", "second"}),
					Tags:     NewTypedListFromStringSlice([]string{"one", "two"}),
				})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Metadata = NewTypedObject(EntryMetadataValue{
					Concepts: NewTypedListFromStringSlice([]string{"second", "first"}),
					Tags:     NewTypedListFromStringSlice([]string{"two", "one"}),
				})

				return value
			}(),
			policy: entryResponseFieldsExact, expectedFields: plan.Fields,
		},
		"content type contradiction": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.ContentTypeID = types.StringValue("remote")

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true, expectedFields: plan.Fields,
		},
		"identity contradiction keeps endpoint identity": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.EntryIdentityModel = NewEntryIdentityModel("remote-space", "remote-environment", "remote-entry")

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true, expectedFields: plan.Fields,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			projected, diags := projectEntryMutationResponse(t.Context(), test.plan, test.response, test.policy)

			assert.Equal(t, test.hasError, diags.HasError(), diags.Errors())
			assert.Equal(t, test.expectedFields, projected.Fields)

			expectedMetadata := test.plan.Metadata
			if test.keepsResponseMetadata {
				expectedMetadata = test.response.Metadata
			}

			assert.Equal(t, expectedMetadata, projected.Metadata)
			assert.Equal(t, test.plan.Timeouts, projected.Timeouts)
			assert.Equal(t, test.plan.ContentTypeID, projected.ContentTypeID)

			assert.Equal(t, test.plan.EntryIdentityModel, projected.EntryIdentityModel)
		})
	}
}

func TestEntryMetadataEquivalent(t *testing.T) {
	t.Parallel()

	metadata := func(concepts, tags []string) TypedObject[EntryMetadataValue] {
		return NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedListFromStringSlice(concepts),
			Tags:     NewTypedListFromStringSlice(tags),
		})
	}

	tests := map[string]struct {
		left, right TypedObject[EntryMetadataValue]
		equivalent  bool
	}{
		"same order":                 {left: metadata([]string{"a"}, []string{"x", "y"}), right: metadata([]string{"a"}, []string{"x", "y"}), equivalent: true},
		"different order":            {left: metadata([]string{"a", "b"}, []string{"x", "y"}), right: metadata([]string{"b", "a"}, []string{"y", "x"}), equivalent: true},
		"different assignments":      {left: metadata([]string{"a"}, []string{"x"}), right: metadata([]string{"b"}, []string{"x"})},
		"duplicates retained":        {left: metadata(nil, []string{"x", "x"}), right: metadata(nil, []string{"x", "x"}), equivalent: true},
		"duplicates significant":     {left: metadata(nil, []string{"x", "x"}), right: metadata(nil, []string{"x", "y"})},
		"null equals null":           {left: NewTypedObjectNull[EntryMetadataValue](), right: NewTypedObjectNull[EntryMetadataValue](), equivalent: true},
		"null differs from unknown":  {left: NewTypedObjectNull[EntryMetadataValue](), right: NewTypedObjectUnknown[EntryMetadataValue]()},
		"null differs from empty":    {left: NewTypedObjectNull[EntryMetadataValue](), right: metadata(nil, nil)},
		"unknown equals unknown":     {left: NewTypedObjectUnknown[EntryMetadataValue](), right: NewTypedObjectUnknown[EntryMetadataValue](), equivalent: true},
		"unknown differs from known": {left: NewTypedObjectUnknown[EntryMetadataValue](), right: metadata(nil, nil)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.equivalent, entryMetadataEquivalent(test.left, test.right))
		})
	}
}

func TestMergeEntryResponseFieldsWithOmissionFallback(t *testing.T) {
	t.Parallel()

	emptyArray := jsontypes.NewNormalizedValue(`{"en-US":[]}`)
	emptyLocales := jsontypes.NewNormalizedValue(`{}`)
	nonemptyArray := jsontypes.NewNormalizedValue(`{"en-US":["value"]}`)
	nullValue := jsontypes.NewNormalizedValue(`{"en-US":null}`)
	jsonNull := jsontypes.NewNormalizedValue(`null`)
	terraformNull := jsontypes.NewNormalizedNull()
	mixedNull := jsontypes.NewNormalizedValue(`{"en-US":[],"de-DE":null}`)
	scalar := jsontypes.NewNormalizedValue(`{"en-US":"value"}`)

	tests := map[string]struct {
		response TypedMap[jsontypes.Normalized]
		fallback TypedMap[jsontypes.Normalized]
		expected TypedMap[jsontypes.Normalized]
	}{
		"known empty top-level map": {
			response: NewTypedMapNull[jsontypes.Normalized](),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
		"missing empty localized array": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"empty": emptyArray}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{"empty": emptyArray}),
		},
		"missing nonempty localized array": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"nonempty": nonemptyArray}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
		"missing empty localized object": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"empty-locales": emptyLocales}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
		"missing localized null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"null": nullValue}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
		"missing JSON null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"optional": jsonNull}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{"optional": jsonNull}),
		},
		"missing Terraform null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"optional": terraformNull}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{"optional": terraformNull}),
		},
		"response value takes precedence over Terraform null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{"optional": scalar}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"optional": terraformNull}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{"optional": scalar}),
		},
		"response value takes precedence over JSON null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{"optional": scalar}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"optional": jsonNull}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{"optional": scalar}),
		},
		"missing mixed empty array and null": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"mixed": mixedNull}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
		"missing scalar": {
			response: NewTypedMap(map[string]jsontypes.Normalized{}),
			fallback: NewTypedMap(map[string]jsontypes.Normalized{"scalar": scalar}),
			expected: NewTypedMap(map[string]jsontypes.Normalized{}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := mergeEntryResponseFieldsWithOmissionFallback(test.response, test.fallback)

			assert.Equal(t, test.expected, actual)
		})
	}
}
