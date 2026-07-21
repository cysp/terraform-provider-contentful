package provider //nolint:testpackage

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveStringSecretConflicts(t *testing.T) {
	t.Parallel()

	_, _, diags := resolveStringSecret(
		types.StringValue("legacy"),
		types.StringValue("write-only"),
		path.Root("value"),
		path.Root("value_wo"),
		true,
	)

	assert.True(t, diags.HasError())
}

func TestWriteOnlySecretHashIncludesPath(t *testing.T) {
	t.Parallel()

	value := types.StringValue("shared-secret")
	valuePath := path.Root("value_wo")
	headerPath := path.Root("headers").AtMapKey(`x-"secret"\key`).AtName("value_wo")

	hash, err := writeOnlySecretHash(valuePath, value)
	require.NoError(t, err)

	assert.True(t, writeOnlySecretHashMatches(valuePath, value, hash))
	assert.False(t, writeOnlySecretHashMatches(headerPath, value, hash))
}

func TestWriteOnlySecretHashesChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	argumentPath := path.Root("value_wo")
	value := types.StringValue("current-secret")

	hash, err := writeOnlySecretHash(argumentPath, value)
	require.NoError(t, err)

	private := fakePrivateState{
		values: map[string][]byte{
			writeOnlySecretHashesPrivateKey: mustJSONMarshal(t, []writeOnlySecretHashRecord{{
				Path: argumentPath.String(),
				Hash: hash,
			}}),
		},
	}

	changed, diags := writeOnlySecretHashesChanged(ctx, private, WriteOnlySecretValues{{
		Path:  argumentPath,
		Value: value,
	}})
	require.False(t, diags.HasError(), diags)
	assert.False(t, changed)

	changed, diags = writeOnlySecretHashesChanged(ctx, private, WriteOnlySecretValues{{
		Path:  argumentPath,
		Value: types.StringValue("rotated-secret"),
	}})
	require.False(t, diags.HasError(), diags)
	assert.True(t, changed)
}

type fakePrivateState struct {
	values map[string][]byte
}

func (s fakePrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return s.values[key], nil
}

func mustJSONMarshal(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return data
}
