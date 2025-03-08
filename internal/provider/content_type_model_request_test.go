package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func TestToOptContentTypeFieldsItemItemsErrorHandling(t *testing.T) {
	t.Parallel()

	itemsObject, itemsObjectDiags := basetypes.NewObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	assert.Empty(t, itemsObjectDiags)

	items, itemsDiags := provider.ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(t.Context(), path.Root("items"), itemsObject)
	assert.NotEmpty(t, itemsDiags)

	assert.EqualValues(t, cm.OptContentTypeRequestFieldsFieldsItemItems{
		Set: false,
	}, items)
}
