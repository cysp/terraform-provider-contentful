//nolint:testpackage // Package-local version projection is the seam under test.
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type taxonomyVersionPrivateData struct{ data map[string][]byte }

func newTaxonomyVersionPrivateData() *taxonomyVersionPrivateData {
	return &taxonomyVersionPrivateData{data: map[string][]byte{}}
}

func (data *taxonomyVersionPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return data.data[key], nil
}

func (data *taxonomyVersionPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	data.data[key] = value

	return nil
}

func TestTaxonomyPriorStateVersionRequiresRefresh(t *testing.T) {
	t.Parallel()

	version, diags := taxonomyPriorStateVersion(t.Context(), newTaxonomyVersionPrivateData())
	assert.Zero(t, version)
	require.True(t, diags.HasError())
	assert.Equal(t, "Taxonomy resource version is unavailable", diags[0].Summary())
}

func TestTaxonomyPriorStateVersionAcceptsZeroAndPositiveVersions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value int
		want  int
	}{
		{name: "zero", value: 0, want: 0},
		{name: "positive", value: 1, want: 1},
		{name: "larger positive", value: 42, want: 42},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			privateData := newTaxonomyVersionPrivateData()
			require.Empty(t, setTaxonomyPrivateVersion(t.Context(), privateData, test.value))

			version, diags := taxonomyPriorStateVersion(t.Context(), privateData)
			require.False(t, diags.HasError())
			assert.Equal(t, test.want, version)
		})
	}
}

func TestTaxonomyPriorStateVersionRejectsInvalidVersions(t *testing.T) {
	t.Parallel()

	privateData := newTaxonomyVersionPrivateData()
	privateData.data[taxonomyVersionPrivateKey] = []byte("-1")

	version, diags := taxonomyPriorStateVersion(t.Context(), privateData)
	assert.Zero(t, version)
	require.True(t, diags.HasError())
	assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
}

func TestTaxonomyPriorStateVersionRejectsMalformedPrivateData(t *testing.T) {
	t.Parallel()

	privateData := newTaxonomyVersionPrivateData()
	privateData.data[taxonomyVersionPrivateKey] = []byte(`"invalid"`)

	version, diags := taxonomyPriorStateVersion(t.Context(), privateData)
	assert.Zero(t, version)
	require.True(t, diags.HasError())
	assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
}

func TestTaxonomyResponseVersionRejectsNegativeVersions(t *testing.T) {
	t.Parallel()

	version, diags := taxonomyResponseVersion(-1)
	assert.Zero(t, version)
	require.True(t, diags.HasError())
	assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
}
