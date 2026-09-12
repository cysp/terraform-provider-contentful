package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func attributeDiagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))

	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok, "expected attribute diagnostic, got %T: %s", diagnostic, diagnostic.Summary())

		paths = append(paths, withPath.Path().String())
	}

	return paths
}

func attributeWarningPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Warnings()))
	for _, diagnostic := range diags.Warnings() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok, "expected attribute diagnostic, got %T: %s", diagnostic, diagnostic.Summary())

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
