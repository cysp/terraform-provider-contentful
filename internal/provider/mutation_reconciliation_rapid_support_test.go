package provider_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// The JSON documents are independently authored representations of the same
// object. The third document changes only array order, which is significant.
// Standard-library encoding supplies the expected response text; no production
// normalization or equivalence operation participates in the oracle.
func drawReconciliationJSON(t *rapid.T) (string, string, string) {
	t.Helper()

	text := rapid.StringOfN(rapid.RuneFrom([]rune("ab <>\"\\\n\t♥")), 0, 12, 12).Draw(t, "JSON text")
	first := rapid.IntRange(-20, 20).Draw(t, "JSON number")
	planned := fmt.Sprintf("{ \"z\": [%d, %d], \"a\": %s }", first, first+1, testJSON(text))
	returned := testJSON(map[string]any{"a": text, "z": []int{first, first + 1}})
	reversed := testJSON(map[string]any{"a": text, "z": []int{first + 1, first}})

	return planned, returned, reversed
}

func requireReconciliationPaths(t *rapid.T, diagnostics diag.Diagnostics, expected ...string) {
	t.Helper()
	require.Len(t, diagnostics, len(expected), "%v", diagnostics)

	if len(diagnostics) == 0 {
		return
	}

	paths := make([]string, 0, len(diagnostics))

	for _, diagnostic := range diagnostics {
		require.Equal(t, diag.SeverityError, diagnostic.Severity())
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok, "%v", diagnostic)

		paths = append(paths, withPath.Path().String())
	}

	assert.Equal(t, expected, paths)
}

// Both values remain present after replacing the first occurrence. This
// distinguishes multiset comparison from a set or length-only comparison.
func drawReconciliationDuplicates(t *rapid.T, first, second string) []string {
	t.Helper()

	firstCount := rapid.IntRange(2, 4).Draw(t, "first multiplicity")
	secondCount := rapid.IntRange(1, 3).Draw(t, "second multiplicity")

	values := make([]string, 0, firstCount+secondCount)
	for range firstCount {
		values = append(values, first)
	}

	for range secondCount {
		values = append(values, second)
	}

	return values
}
