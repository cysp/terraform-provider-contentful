package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToOptContentTypeFieldsItemItemsErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	_, itemsObjectDiags := NewTypedObjectFromAttributes[ContentTypeFieldItemsValue](ctx, map[string]attr.Value{
		"type":        types.StringNull(),
		"link_type":   types.StringNull(),
		"validations": NewTypedList([]types.String{types.StringNull()}),
	})
	assert.NotEmpty(t, itemsObjectDiags)
}

func TestContentTypeFieldsRejectNullAndUnknownObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]TypedObject[ContentTypeFieldValue]{
		"null":    NewTypedObjectNull[ContentTypeFieldValue](),
		"unknown": NewTypedObjectUnknown[ContentTypeFieldValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := FieldsListToContentTypeRequestDataFields(
				t.Context(),
				path.Root("fields"),
				NewTypedList([]TypedObject[ContentTypeFieldValue]{value}),
			)
			assert.Nil(t, result)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{"fields[0]"}, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeAllowedResourceRequiresExactlyOneAlternative(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("allowed_resources").AtListIndex(0)
	external := NewTypedObject(ContentTypeFieldAllowedResourceItemExternalValue{
		TypeID: types.StringValue("ExternalType"),
	})
	contentfulEntry := NewTypedObject(ContentTypeFieldAllowedResourceItemContentfulEntryValue{
		Source:       types.StringValue("crn:contentful:::content:spaces/example"),
		ContentTypes: NewTypedList([]types.String{types.StringValue("article")}),
	})

	tests := map[string]struct {
		value         ContentTypeFieldAllowedResourceItemValue
		expected      cm.ResourceLink
		expectedPaths []string
	}{
		"external": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External:        external,
				ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			},
			expected: cm.ResourceLink{Type: "ExternalType"},
		},
		"contentful entry": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External:        NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
				ContentfulEntry: contentfulEntry,
			},
			expected: cm.ResourceLink{
				Type:         "Contentful:Entry",
				Source:       cm.NewOptString("crn:contentful:::content:spaces/example"),
				ContentTypes: []string{"article"},
			},
		},
		"neither": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External:        NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
				ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			},
			expectedPaths: []string{"allowed_resources[0]"},
		},
		"both": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External:        external,
				ContentfulEntry: contentfulEntry,
			},
			expectedPaths: []string{"allowed_resources[0]"},
		},
		"unknown": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External:        NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemExternalValue](),
				ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			},
			expectedPaths: []string{"allowed_resources[0].external"},
		},
		"invalid external": {
			value: ContentTypeFieldAllowedResourceItemValue{
				External: NewTypedObject(ContentTypeFieldAllowedResourceItemExternalValue{
					TypeID: types.StringUnknown(),
				}),
				ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			},
			expectedPaths: []string{"allowed_resources[0].external.type"},
		},
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := value.value.ToResourceLink(t.Context(), valuePath)

			assert.Equal(t, value.expected, actual)
			assert.ElementsMatch(t, value.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeValidationsFailClosedWithExactPath(t *testing.T) {
	t.Parallel()

	actual, diags := ValidationsListToContentTypeRequestDataFieldValidations(
		t.Context(),
		path.Root("fields").AtListIndex(0).AtName("validations"),
		NewTypedList([]jsontypes.Normalized{
			jsontypes.NewNormalizedValue(`{"size":{"min":1}}`),
			jsontypes.NewNormalizedNull(),
		}),
	)

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"fields[0].validations[1]"}, diagnosticPaths(t, diags))
}
