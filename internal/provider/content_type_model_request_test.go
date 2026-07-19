package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
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
