package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToOptContentTypeFieldsItemItemsErrorHandling(t *testing.T) {
	t.Parallel()

	itemsObject, itemsObjectDiags := provider.NewContentTypeFieldItemsValueKnownFromAttributes(t.Context(), map[string]attr.Value{
		"type":        types.StringNull(),
		"link_type":   types.StringNull(),
		"validations": types.ListValueMust(types.BoolType, []attr.Value{types.BoolNull()}),
	})
	assert.Empty(t, itemsObjectDiags)

	items, itemsDiags := provider.ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(t.Context(), path.Root("items"), itemsObject)
	assert.NotEmpty(t, itemsDiags)

	assert.EqualValues(t, cm.OptContentTypeRequestFieldsFieldsItemItems{
		Value: cm.ContentTypeRequestFieldsFieldsItemItems{
			Validations: []jx.Raw{},
		},
		Set: true,
	}, items)

	assert.NotEmpty(t, itemsDiags.Errors())
}
