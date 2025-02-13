package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestReadWebhookDefinitionFilterTermString(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`"abc"`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`123`),
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := provider.ReadWebhookDefinitionFilterTermString(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func TestReadWebhookDefinitionFilterTermStringArray(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`["abc"]`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`"abc"`),
			expectError: true,
		},
		"wrong element type": {
			input:       []byte(`["abc",123]`),
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := provider.ReadWebhookDefinitionFilterTermStringArray(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func TestReadWebhookDefinitionFilterTermStringObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		name        string
		expectNull  bool
		expectValue string
		expectError bool
	}{
		"valid": {
			input:       []byte(`{"value":"abc"}`),
			name:        "value",
			expectValue: "abc",
		},
		"valid with excess": {
			input:       []byte(`{"a":"b","value":"abc","c":"d"}`),
			name:        "value",
			expectValue: "abc",
		},
		"value absent": {
			input:      []byte(`{"a":"b"}`),
			name:       "value",
			expectNull: true,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`123`),
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := provider.ReadWebhookDefinitionFilterTermStringObject(
				ctx,
				path.Root("test"),
				testcase.name,
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}

			if testcase.expectNull {
				assert.True(t, value.IsNull())
			} else {
				assert.Equal(t, testcase.expectValue, value.ValueString())
			}
		})
	}
}
