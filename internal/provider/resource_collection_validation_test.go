package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionNullValidatorsRejectNullChildrenAndDeferUnknowns(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	contentTypeFields, ok := ContentTypeResourceSchema(ctx).Attributes["fields"].(schema.ListNestedAttribute)
	require.True(t, ok)

	fieldType := NewTypedObjectNull[ContentTypeFieldValue]().Type(ctx)
	validateListAttribute(
		t,
		contentTypeFields.Validators,
		path.Root("fields"),
		types.ListValueMust(fieldType, []attr.Value{NewTypedObjectNull[ContentTypeFieldValue]()}),
		types.ListUnknown(fieldType),
	)

	membershipRoles, ok := TeamSpaceMembershipResourceSchema(ctx).Attributes["roles"].(schema.ListAttribute)
	require.True(t, ok)
	validateListAttribute(
		t,
		membershipRoles.Validators,
		path.Root("roles"),
		types.ListValueMust(types.StringType, []attr.Value{types.StringNull()}),
		types.ListUnknown(types.StringType),
	)

	marketplace, ok := AppInstallationResourceSchema(ctx).Attributes["marketplace"].(schema.SetAttribute)
	require.True(t, ok)

	for _, value := range []struct {
		name          string
		config        types.Set
		expectedError bool
	}{
		{name: "null child", config: types.SetValueMust(types.StringType, []attr.Value{types.StringNull()}), expectedError: true},
		{name: "unknown container", config: types.SetUnknown(types.StringType)},
	} {
		t.Run("marketplace "+value.name, func(t *testing.T) {
			t.Parallel()

			response := validator.SetResponse{}
			for _, itemValidator := range marketplace.Validators {
				itemValidator.ValidateSet(ctx, validator.SetRequest{Path: path.Root("marketplace"), ConfigValue: value.config}, &response)
			}

			assert.Equal(t, value.expectedError, response.Diagnostics.HasError(), response.Diagnostics)
		})
	}
}

func validateListAttribute(
	t *testing.T,
	validators []validator.List,
	valuePath path.Path,
	nullChild types.List,
	unknownContainer types.List,
) {
	t.Helper()
	require.NotEmpty(t, validators)

	nullResponse := validator.ListResponse{}
	for _, itemValidator := range validators {
		itemValidator.ValidateList(t.Context(), validator.ListRequest{Path: valuePath, ConfigValue: nullChild}, &nullResponse)
	}

	assert.True(t, nullResponse.Diagnostics.HasError())

	unknownResponse := validator.ListResponse{}
	for _, itemValidator := range validators {
		itemValidator.ValidateList(t.Context(), validator.ListRequest{Path: valuePath, ConfigValue: unknownContainer}, &unknownResponse)
	}

	assert.False(t, unknownResponse.Diagnostics.HasError(), unknownResponse.Diagnostics)
}
