package provider_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTaxonomyEmptyPatch(t *testing.T, body []byte, paths []string) {
	t.Helper()

	var patch []struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
	require.NoError(t, json.Unmarshal(body, &patch))
	require.Len(t, patch, len(paths))

	for index, wantPath := range paths {
		assert.Equal(t, "add", patch[index].Op)
		assert.Equal(t, wantPath, patch[index].Path)
		assert.JSONEq(t, `{}`, string(patch[index].Value))
	}
}
