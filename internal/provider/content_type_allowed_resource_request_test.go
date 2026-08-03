package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
			expected:      cm.ResourceLink{Type: "ExternalType"},
			expectedPaths: []string{},
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
			expectedPaths: []string{},
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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.value.ToResourceLink(t.Context(), valuePath)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeAllowedContentfulEntryRejectsUnresolvedValues(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("allowed_resources").AtListIndex(0).AtName("contentful_entry")
	tests := map[string]struct {
		value         ContentTypeFieldAllowedResourceItemContentfulEntryValue
		expectedPaths []string
	}{
		"unknown source": {
			value: ContentTypeFieldAllowedResourceItemContentfulEntryValue{
				Source:       types.StringUnknown(),
				ContentTypes: NewTypedList([]types.String{}),
			},
			expectedPaths: []string{"allowed_resources[0].contentful_entry.source"},
		},
		"null source": {
			value: ContentTypeFieldAllowedResourceItemContentfulEntryValue{
				Source:       types.StringNull(),
				ContentTypes: NewTypedList([]types.String{}),
			},
			expectedPaths: []string{"allowed_resources[0].contentful_entry.source"},
		},
		"unknown content types": {
			value: ContentTypeFieldAllowedResourceItemContentfulEntryValue{
				Source:       types.StringValue("source"),
				ContentTypes: NewTypedListUnknown[types.String](),
			},
			expectedPaths: []string{"allowed_resources[0].contentful_entry.content_types"},
		},
		"null content types": {
			value: ContentTypeFieldAllowedResourceItemContentfulEntryValue{
				Source:       types.StringValue("source"),
				ContentTypes: NewTypedListNull[types.String](),
			},
			expectedPaths: []string{"allowed_resources[0].contentful_entry.content_types"},
		},
		"unresolved content type elements": {
			value: ContentTypeFieldAllowedResourceItemContentfulEntryValue{
				Source: types.StringValue("source"),
				ContentTypes: NewTypedList([]types.String{
					types.StringValue("valid"),
					types.StringNull(),
					types.StringUnknown(),
				}),
			},
			expectedPaths: []string{
				"allowed_resources[0].contentful_entry.content_types[1]",
				"allowed_resources[0].contentful_entry.content_types[2]",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.value.ToResourceLink(t.Context(), valuePath)

			assert.Equal(t, cm.ResourceLink{}, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
		})
	}
}
