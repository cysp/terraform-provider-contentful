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

func TestNewEntryFieldsFromResponseRejectsWholeFieldJSONNull(t *testing.T) {
	t.Parallel()
	_, diags := NewEntryFieldsFromResponse(t.Context(), path.Root("fields"), cm.NewOptEntryFields(cm.EntryFields{"optional": jx.Raw("null")}))
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`fields["optional"]`}, attributeDiagnosticPaths(t, diags))
}

func TestMergeEntryFieldsWithFallback(t *testing.T) {
	t.Parallel()

	returnedValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("returned"))})
	configuredValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("configured"))})
	missingValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("missing"))})
	returned := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"returned": returnedValue})
	configured := NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
		"returned": configuredValue,
		"missing":  missingValue,
	})

	actual := mergeEntryFieldsWithFallback(returned, configured)

	assert.Equal(t, map[string]TypedMap[jsontypes.Normalized]{
		"returned": returnedValue,
		"missing":  missingValue,
	}, actual.Elements())
	assert.Equal(t, map[string]TypedMap[jsontypes.Normalized]{"returned": returnedValue}, returned.Elements())
	assert.Equal(t, map[string]TypedMap[jsontypes.Normalized]{
		"returned": configuredValue,
		"missing":  missingValue,
	}, configured.Elements())
}

func TestMergeEntryFieldsWithFallbackInitializesNullResponse(t *testing.T) {
	t.Parallel()

	value := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("configured"))})
	actual := mergeEntryFieldsWithFallback(
		NewTypedMapNull[TypedMap[jsontypes.Normalized]](),
		NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"field": value}),
	)

	assert.False(t, actual.IsNull())
	assert.Equal(t, map[string]TypedMap[jsontypes.Normalized]{"field": value}, actual.Elements())
}

func TestProjectEntryMutationResponse(t *testing.T) {
	t.Parallel()

	managed := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("managed"))})
	external := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("external"))})
	changed := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("changed"))})
	defaulted := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("default"))})
	terraformNull := NewTypedMapNull[jsontypes.Normalized]()
	emptyArray := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON([]any{}))})
	semanticPlanValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(`{"first":"one","second":"two"}`)})
	semanticResponseValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(`{ "second": "two", "first": "one" }`)})
	require.NotEqual(t, semanticPlanValue.Elements()["en-US"].ValueString(), semanticResponseValue.Elements()["en-US"].ValueString())

	remoteMetadata := NewTypedObject[EntryMetadataValue](EntryMetadataValue{
		Concepts: NewTypedListFromStringSlice([]string{"remote"}),
		Tags:     NewTypedListFromStringSlice([]string{}),
	})
	plan := EntryModel{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID("space", "environment", "entry"),
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("article"),
		PublishedVersion:   types.Int64Null(),
		Fields:             NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed}),
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
		expectedFields        TypedMap[TypedMap[jsontypes.Normalized]]
		keepsResponseMetadata bool
	}{
		"semantically equal JSON": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": semanticPlanValue})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": semanticResponseValue})

				return value
			}(),
			policy:         entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": semanticPlanValue}),
		},
		"whole fields member omitted": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMapNull[TypedMap[jsontypes.Normalized]]()

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMapNull[TypedMap[jsontypes.Normalized]](),
		},
		"present partial response missing nonempty field": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "external": external})

				return value
			}(),
			response: plan, policy: entryResponseFieldsCreationDefaults, hasError: true, expectedFields: plan.Fields,
		},
		"present partial response restores all-empty localized array": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "empty": emptyArray})

				return value
			}(),
			response: plan, policy: entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "empty": emptyArray}),
		},
		"additional response field accepted": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsCreationDefaults, expectedFields: plan.Fields,
		},
		"response default for omitted Terraform null accepted on create": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "default": terraformNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsCreationDefaults,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
				"managed": managed, "default": terraformNull,
			}),
		},
		"omitted Terraform null accepted by exact response": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "optional": terraformNull})

				return value
			}(),
			response: plan, policy: entryResponseFieldsExact,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
				"managed": managed, "optional": terraformNull,
			}),
		},
		"response value for omitted Terraform null rejected by exact response": {
			plan: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "optional": terraformNull})

				return value
			}(),
			response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "optional": defaulted})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{
				"managed": managed, "optional": defaulted,
			}),
		},
		"additional response field rejected by exact policy": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "default": defaulted})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": managed, "default": defaulted}),
		},
		"changed managed field": {
			plan: plan, response: func() EntryModel {
				value := plan
				value.Fields = NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": changed})

				return value
			}(),
			policy: entryResponseFieldsExact, hasError: true,
			expectedFields: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"managed": changed}),
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

			projected, diags, _ := projectEntryMutationResponse(t.Context(), test.plan, test.response, test.policy)

			assert.Equal(t, test.hasError, diags.HasError(), diags.Errors())
			assert.Equal(t, test.expectedFields, projected.Fields)

			expectedMetadata := test.plan.Metadata
			if test.keepsResponseMetadata {
				expectedMetadata = test.response.Metadata
			}

			assert.Equal(t, expectedMetadata, projected.Metadata)
			assert.Equal(t, test.plan.Timeouts, projected.Timeouts)
			assert.Equal(t, test.response.ContentTypeID, projected.ContentTypeID)

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

	emptyArray := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON([]any{}))})
	emptyLocales := NewTypedMap(map[string]jsontypes.Normalized{})
	nonemptyArray := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON([]any{"value"}))})
	nullValue := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue("null")})
	terraformNull := NewTypedMapNull[jsontypes.Normalized]()
	mixedNull := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue("[]"), "de-DE": jsontypes.NewNormalizedValue("null")})
	scalar := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue(testJSON("value"))})

	emptyWithOmittedLocale := NewTypedMap(map[string]jsontypes.Normalized{"en-US": jsontypes.NewNormalizedValue("[]"), "de-DE": jsontypes.NewNormalizedNull()})
	onlyOmittedLocale := NewTypedMap(map[string]jsontypes.Normalized{"de-DE": jsontypes.NewNormalizedNull()})

	tests := map[string]struct {
		response TypedMap[TypedMap[jsontypes.Normalized]]
		fallback TypedMap[TypedMap[jsontypes.Normalized]]
		expected TypedMap[TypedMap[jsontypes.Normalized]]
	}{
		"known empty top-level map": {
			response: NewTypedMapNull[TypedMap[jsontypes.Normalized]](),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"empty array and omitted locale": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"field": emptyWithOmittedLocale}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"field": emptyWithOmittedLocale}),
		},
		"only omitted locales": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"field": onlyOmittedLocale}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"missing empty localized array": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"empty": emptyArray}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"empty": emptyArray}),
		},
		"missing nonempty localized array": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"nonempty": nonemptyArray}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"missing empty localized object": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"empty-locales": emptyLocales}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"missing localized null": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"null": nullValue}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"missing Terraform null": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"optional": terraformNull}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"optional": terraformNull}),
		},
		"response value takes precedence over Terraform null": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"optional": scalar}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"optional": terraformNull}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"optional": scalar}),
		},
		"missing mixed empty array and null": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"mixed": mixedNull}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
		},
		"missing scalar": {
			response: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
			fallback: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{"scalar": scalar}),
			expected: NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
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
