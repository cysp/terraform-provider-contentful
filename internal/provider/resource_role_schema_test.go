package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleSchemaRejectsAllCombinedWithOtherValues(t *testing.T) {
	t.Parallel()

	permissionsAttribute, policyActionsAttribute := roleSchemaValueAttributes(t)
	permissionPath := path.Root("permissions").AtMapKey("ContentModel")
	permissionDiags := validateRolePermissionValuesConfig(t, permissionsAttribute, types.MapValueMust(
		permissionsAttribute.ElementType,
		map[string]attr.Value{"ContentModel": NewTypedListFromStringSlice([]string{"all", "read"})},
	))
	assertRoleValidationDiagnostic(
		t,
		permissionDiags,
		permissionPath,
		"Invalid permission values",
		`"all" must be specified by itself. Remove "all" or the other permission values from this list.`,
	)

	policyActionsPath := path.Root("policies").AtListIndex(0).AtName("actions")
	policyActionsDiags := validateRolePolicyActionsConfig(t, policyActionsAttribute, types.ListValueMust(
		types.StringType,
		[]attr.Value{types.StringValue("all"), types.StringValue("read")},
	))
	assertRoleValidationDiagnostic(
		t,
		policyActionsDiags,
		policyActionsPath,
		"Invalid policy actions",
		`"all" must be specified by itself. Remove "all" or the other policy actions from this list.`,
	)
}

func TestRoleAllValidatorsHandleUnresolvedValues(t *testing.T) {
	t.Parallel()

	permissionsAttribute, policyActionsAttribute := roleSchemaValueAttributes(t)

	assert.Empty(t, validateRolePermissionValuesConfig(t, permissionsAttribute, types.MapNull(permissionsAttribute.ElementType)))
	assert.Empty(t, validateRolePermissionValuesConfig(t, permissionsAttribute, types.MapUnknown(permissionsAttribute.ElementType)))
	assert.Empty(t, validateRolePermissionValuesConfig(t, permissionsAttribute, types.MapValueMust(
		permissionsAttribute.ElementType,
		map[string]attr.Value{
			"ContentModel": NewTypedList([]types.String{types.StringValue("read"), types.StringUnknown()}),
		},
	)))

	assert.Empty(t, validateRolePolicyActionsConfig(t, policyActionsAttribute, types.ListNull(types.StringType)))
	assert.Empty(t, validateRolePolicyActionsConfig(t, policyActionsAttribute, types.ListUnknown(types.StringType)))
	assert.Empty(t, validateRolePolicyActionsConfig(t, policyActionsAttribute, types.ListValueMust(
		types.StringType,
		[]attr.Value{types.StringValue("read"), types.StringUnknown()},
	)))
}

func TestRoleSchemaRejectsKnownAllWithUnknownSibling(t *testing.T) {
	t.Parallel()

	permissionsAttribute, policyActionsAttribute := roleSchemaValueAttributes(t)
	permissionPath := path.Root("permissions").AtMapKey("ContentModel")
	permissionDiags := validateRolePermissionValuesConfig(t, permissionsAttribute, types.MapValueMust(
		permissionsAttribute.ElementType,
		map[string]attr.Value{
			"ContentModel": NewTypedList([]types.String{types.StringValue("all"), types.StringUnknown()}),
		},
	))
	assertRoleValidationDiagnostic(
		t,
		permissionDiags,
		permissionPath,
		"Invalid permission values",
		`"all" must be specified by itself. Remove "all" or the other permission values from this list.`,
	)

	policyActionsPath := path.Root("policies").AtListIndex(0).AtName("actions")
	policyActionsDiags := validateRolePolicyActionsConfig(t, policyActionsAttribute, types.ListValueMust(
		types.StringType,
		[]attr.Value{types.StringUnknown(), types.StringValue("all")},
	))
	assertRoleValidationDiagnostic(
		t,
		policyActionsDiags,
		policyActionsPath,
		"Invalid policy actions",
		`"all" must be specified by itself. Remove "all" or the other policy actions from this list.`,
	)
}

func roleSchemaValueAttributes(t *testing.T) (schema.MapAttribute, schema.ListAttribute) {
	t.Helper()

	resourceSchema := RoleResourceSchema(t.Context())
	permissionsAttribute, ok := resourceSchema.Attributes["permissions"].(schema.MapAttribute)
	require.True(t, ok)
	policiesAttribute, ok := resourceSchema.Attributes["policies"].(schema.ListNestedAttribute)
	require.True(t, ok)
	policyActionsAttribute, ok := policiesAttribute.NestedObject.Attributes["actions"].(schema.ListAttribute)
	require.True(t, ok)

	return permissionsAttribute, policyActionsAttribute
}

func validateRolePermissionValuesConfig(t *testing.T, attribute schema.MapAttribute, value types.Map) diag.Diagnostics {
	t.Helper()

	response := validator.MapResponse{}
	for _, valueValidator := range attribute.Validators {
		valueValidator.ValidateMap(t.Context(), validator.MapRequest{Path: path.Root("permissions"), ConfigValue: value}, &response)
	}

	return response.Diagnostics
}

func validateRolePolicyActionsConfig(t *testing.T, attribute schema.ListAttribute, value types.List) diag.Diagnostics {
	t.Helper()

	response := validator.ListResponse{}
	for _, valueValidator := range attribute.Validators {
		valueValidator.ValidateList(t.Context(), validator.ListRequest{
			Path: path.Root("policies").AtListIndex(0).AtName("actions"), ConfigValue: value,
		}, &response)
	}

	return response.Diagnostics
}

func assertRoleValidationDiagnostic(t *testing.T, diagnostics diag.Diagnostics, expectedPath path.Path, summary, detail string) {
	t.Helper()

	require.Len(t, diagnostics, 1)
	assert.Equal(t, diag.SeverityError, diagnostics[0].Severity())
	assert.Equal(t, summary, diagnostics[0].Summary())
	assert.Equal(t, detail, diagnostics[0].Detail())
	diagnostic, ok := diagnostics[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, expectedPath, diagnostic.Path())
}
