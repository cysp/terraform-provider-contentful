package provider

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise the internal traversal interface with typed pages so
// projection diagnostics can be checked independently of generated decoding.
func TestReadContentfulCollectionPreservesDiagnostics(t *testing.T) {
	t.Parallel()

	pageWarning := diag.NewWarningDiagnostic("Page warning", "The first page included a warning.")
	firstWarning := diag.NewAttributeWarningDiagnostic(path.Root("items").AtListIndex(0), "Item warning", "The first item included a warning.")
	lastWarning := diag.NewAttributeWarningDiagnostic(path.Root("items").AtListIndex(2), "Item warning", "The final item included a warning.")

	var offsets []int64

	var indexes []int

	items, diagnostics := readContentfulCollection(t.Context(), "Failed to read test collection",
		func(_ context.Context, skip int64) (contentfulCollection[cm.Space], diag.Diagnostics) {
			offsets = append(offsets, skip)
			if skip == 0 {
				return &cm.SpaceCollection{Total: cm.NewOptInt(3), Items: []cm.Space{{Name: "duplicate"}, {Name: "duplicate"}}}, diag.Diagnostics{pageWarning}
			}

			return &cm.SpaceCollection{Total: cm.NewOptInt(3), Items: []cm.Space{{Name: "last"}}}, nil
		},
		func(index int, item cm.Space) (string, diag.Diagnostics) {
			indexes = append(indexes, index)

			switch index {
			case 0:
				return item.Name, diag.Diagnostics{firstWarning}
			case 2:
				return item.Name, diag.Diagnostics{lastWarning}
			default:
				return item.Name, nil
			}
		},
	)

	assert.Equal(t, []int64{0, 2}, offsets)
	assert.Equal(t, []int{0, 1, 2}, indexes)
	assert.Equal(t, []string{"duplicate", "duplicate", "last"}, items)
	assert.Equal(t, diag.Diagnostics{pageWarning, firstWarning, lastWarning}, diagnostics)
}

func TestReadContentfulCollectionErrorRetainsEarlierDiagnostics(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		fetchFail bool
	}{
		{name: "fetch error", fetchFail: true},
		{name: "projection error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			pageWarning := diag.NewWarningDiagnostic("Page warning", "The first page included a warning.")
			itemWarning := diag.NewAttributeWarningDiagnostic(path.Root("items").AtListIndex(0), "Item warning", "The first item included a warning.")
			failure := diag.NewErrorDiagnostic("Collection failure", "The next page cannot be represented.")

			var offsets []int64

			var projected []string

			items, diagnostics := readContentfulCollection(t.Context(), "Failed to read test collection",
				func(_ context.Context, skip int64) (contentfulCollection[cm.Space], diag.Diagnostics) {
					offsets = append(offsets, skip)
					if skip == 0 {
						return &cm.SpaceCollection{Total: cm.NewOptInt(3), Items: []cm.Space{{Name: "first"}}}, diag.Diagnostics{pageWarning}
					}

					if test.fetchFail {
						return nil, diag.Diagnostics{failure}
					}

					return &cm.SpaceCollection{Total: cm.NewOptInt(3), Items: []cm.Space{{Name: "invalid"}, {Name: "unvisited"}}}, nil
				},
				func(_ int, item cm.Space) (string, diag.Diagnostics) {
					projected = append(projected, item.Name)
					if item.Name == "first" {
						return item.Name, diag.Diagnostics{itemWarning}
					}

					return "", diag.Diagnostics{failure}
				},
			)

			assert.Nil(t, items)
			assert.Equal(t, diag.Diagnostics{pageWarning, itemWarning, failure}, diagnostics)
			assert.Equal(t, []int64{0, 1}, offsets)

			if test.fetchFail {
				assert.Equal(t, []string{"first"}, projected)
			} else {
				assert.Equal(t, []string{"first", "invalid"}, projected)
			}
		})
	}
}

func TestReadContentfulCollectionCancellationRetainsEarlierDiagnostics(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	warning := diag.NewWarningDiagnostic("Page warning", "The first page included a warning.")

	var offsets []int64

	items, diagnostics := readContentfulCollection(ctx, "Failed to read test collection",
		func(_ context.Context, skip int64) (contentfulCollection[cm.Space], diag.Diagnostics) {
			offsets = append(offsets, skip)

			return &cm.SpaceCollection{Items: []cm.Space{{Name: "first"}}}, diag.Diagnostics{warning}
		},
		func(_ int, item cm.Space) (string, diag.Diagnostics) {
			cancel()

			return item.Name, nil
		},
	)

	assert.Nil(t, items)
	assert.Equal(t, []int64{0}, offsets)
	require.Len(t, diagnostics, 2)
	assert.Equal(t, warning, diagnostics[0])
	assert.Equal(t, diag.NewErrorDiagnostic("Failed to read test collection", context.Canceled.Error()), diagnostics[1])
}
