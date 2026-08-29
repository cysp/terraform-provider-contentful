//nolint:testpackage // The focused diagnostic adapter is intentionally private production behavior.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingLifecycleDraftMatchesCheckpoint(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		markedVersion    int
		version          int
		publishedVersion cm.OptInt
		checkpointed     types.Int64
		expected         bool
	}{
		"unpublished exact draft":       {markedVersion: 3, version: 3, checkpointed: types.Int64Null(), expected: true},
		"older publication":             {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(1), checkpointed: types.Int64Value(1), expected: true},
		"zero publication":              {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(0), checkpointed: types.Int64Value(0), expected: true},
		"older publication disappeared": {markedVersion: 3, version: 3, checkpointed: types.Int64Value(1)},
		"older publication changed":     {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(2), checkpointed: types.Int64Value(1)},
		"negative publication":          {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(-1), checkpointed: types.Int64Value(-1)},
		"publication equals version":    {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(3), checkpointed: types.Int64Value(3)},
		"future publication":            {markedVersion: 3, version: 3, publishedVersion: cm.NewOptInt(4), checkpointed: types.Int64Value(4)},
		"superseded draft":              {markedVersion: 3, version: 4, publishedVersion: cm.NewOptInt(1), checkpointed: types.Int64Value(1)},
		"unknown checkpoint":            {markedVersion: 3, version: 3, checkpointed: types.Int64Unknown()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, pendingLifecycleDraftMatchesCheckpoint(
				test.markedVersion, test.version, test.publishedVersion, test.checkpointed,
			))
		})
	}
}

func TestAppendDiagnosticsWithErrorsAsWarningsPreservesAttributePath(t *testing.T) {
	t.Parallel()

	publishedVersionPath := path.Root("published_version")
	source := diag.Diagnostics{}
	source.AddAttributeError(publishedVersionPath, "Publication contradiction", "The tuple was contradictory.")

	destination := diag.Diagnostics{}

	appendDiagnosticsWithErrorsAsWarnings(&destination, source, "Terraform retained authority.")

	require.Len(t, destination, 1)
	assert.Equal(t, diag.SeverityWarning, destination[0].Severity())
	diagnosticWithPath, ok := destination[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, publishedVersionPath, diagnosticWithPath.Path())
	assert.Equal(t, "Publication contradiction", diagnosticWithPath.Summary())
	assert.Equal(t, "The tuple was contradictory. Terraform retained authority.", diagnosticWithPath.Detail())
}
