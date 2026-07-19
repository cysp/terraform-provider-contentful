package provider_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTestConversion = errors.New("test conversion failed")

func TestConvertKnownObjectListElementsFailsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	elements := []TypedObject[WebhookHeaderValue]{
		NewTypedObject(WebhookHeaderValue{Value: types.StringValue("first"), Secret: types.BoolValue(false)}),
		NewTypedObjectNull[WebhookHeaderValue](),
		NewTypedObjectUnknown[WebhookHeaderValue](),
		NewTypedObject(WebhookHeaderValue{Value: types.StringValue("last"), Secret: types.BoolValue(false)}),
	}

	result, diags := ConvertKnownObjectListElements(
		t.Context(),
		path.Root("items"),
		elements,
		func(_ context.Context, itemPath path.Path, value WebhookHeaderValue) (string, diag.Diagnostics) {
			if value.Value.ValueString() == "last" {
				return "", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(itemPath, "Conversion failed", errTestConversion.Error())}
			}

			return value.Value.ValueString(), nil
		},
	)

	assert.Nil(t, result)
	require.Len(t, diags.Errors(), 3)
	assert.Equal(t, []string{"items[1]", "items[2]", "items[3]"}, diagnosticPaths(t, diags))
}

func TestConvertKnownObjectMapElementsOrdersKeys(t *testing.T) {
	t.Parallel()

	elements := map[string]TypedObject[WebhookHeaderValue]{
		"z": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("last"), Secret: types.BoolValue(false)}),
		"a": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("first"), Secret: types.BoolValue(false)}),
	}

	result, diags := ConvertKnownObjectMapElements(
		t.Context(),
		path.Root("items"),
		elements,
		func(_ context.Context, _ path.Path, key string, _ WebhookHeaderValue) (string, diag.Diagnostics) {
			return key, nil
		},
	)

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, []string{"a", "z"}, result)
}

func diagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
