package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToOptContentTypeFieldsItemItemsErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	_, itemsObjectDiags := provider.NewContentTypeFieldItemsValueKnownFromAttributes(ctx, map[string]attr.Value{
		"type":        types.StringNull(),
		"link_type":   types.StringNull(),
		"validations": DiagsNoErrorsMust(provider.NewTypedList(ctx, []types.String{types.StringNull()})),
	})
	assert.NotEmpty(t, itemsObjectDiags)
}
