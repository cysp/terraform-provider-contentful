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

func TestTaxonomyPrivateVersionDistinguishesMissingData(t *testing.T) {
	t.Parallel()

	version, found, diags := taxonomyPrivateVersion(t.Context(), newTaxonomyVersionPrivateData())
	require.False(t, diags.HasError())
	assert.False(t, found)
	assert.Zero(t, version)
}

func TestTaxonomyPrivateVersionAcceptsStoredPositiveVersions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value int
		want  int
	}{
		{name: "positive", value: 1, want: 1},
		{name: "larger positive", value: 42, want: 42},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			privateData := newTaxonomyVersionPrivateData()
			require.Empty(t, setTaxonomyPrivateVersion(t.Context(), privateData, test.value))

			version, found, diags := taxonomyPrivateVersion(t.Context(), privateData)
			require.False(t, diags.HasError())
			assert.True(t, found)
			assert.Equal(t, test.want, version)
		})
	}
}

func TestTaxonomyPrivateVersionRejectsNonpositiveStoredVersions(t *testing.T) {
	t.Parallel()

	for _, encoded := range []string{"0", "-1"} {
		t.Run(encoded, func(t *testing.T) {
			t.Parallel()

			privateData := newTaxonomyVersionPrivateData()
			privateData.data[taxonomyVersionPrivateKey] = []byte(encoded)

			version, found, diags := taxonomyPrivateVersion(t.Context(), privateData)
			assert.Zero(t, version)
			assert.True(t, found)
			require.True(t, diags.HasError())
			assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
		})
	}
}

func TestTaxonomyPrivateVersionRejectsMalformedPrivateData(t *testing.T) {
	t.Parallel()

	privateData := newTaxonomyVersionPrivateData()
	privateData.data[taxonomyVersionPrivateKey] = []byte(`"invalid"`)

	version, found, diags := taxonomyPrivateVersion(t.Context(), privateData)
	assert.Zero(t, version)
	assert.True(t, found)
	require.True(t, diags.HasError())
	assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
}

func TestTaxonomyResponseVersionRejectsNonpositiveVersions(t *testing.T) {
	t.Parallel()

	for _, responseVersion := range []int{0, -1} {
		version, diags := taxonomyResponseVersion(responseVersion)
		assert.Zero(t, version)
		require.True(t, diags.HasError())
		assert.Equal(t, "Invalid taxonomy resource version", diags[0].Summary())
	}
}

func TestTaxonomyDeleteResponseVersionRequiresRequestedIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name           string
		organizationID string
		resourceID     string
	}{
		{name: "organization differs", organizationID: "other-organization", resourceID: "resource-id"},
		{name: "resource differs", organizationID: "organization-id", resourceID: "other-resource"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			version, diags := taxonomyDeleteResponseVersion("taxonomy resource", "organization-id", "resource-id", test.organizationID, test.resourceID, 1)
			assert.Zero(t, version)
			require.True(t, diags.HasError())
			assert.Equal(t, "Unexpected Contentful taxonomy resource response", diags[0].Summary())
		})
	}
}

func TestTaxonomyDeleteResponseVersionAcceptsPositiveVersion(t *testing.T) {
	t.Parallel()

	version, diags := taxonomyDeleteResponseVersion("taxonomy resource", "organization-id", "resource-id", "organization-id", "resource-id", 42)
	require.False(t, diags.HasError())
	assert.Equal(t, 42, version)
}
