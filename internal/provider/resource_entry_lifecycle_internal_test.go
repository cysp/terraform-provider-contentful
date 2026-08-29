package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryPublicationResponseTuple(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		responseVersion  int
		publishedVersion cm.OptInt
		severity         diag.Severity
	}{
		"exact create response": {
			responseVersion: 4, publishedVersion: cm.NewOptInt(3),
		},
		"missing published version": {
			responseVersion: 4, severity: diag.SeverityError,
		},
		"wrong published version": {
			responseVersion: 4, publishedVersion: cm.NewOptInt(2), severity: diag.SeverityError,
		},
		"version did not advance": {
			responseVersion: 3, publishedVersion: cm.NewOptInt(3), severity: diag.SeverityError,
		},
		"version regressed": {
			responseVersion: 2, publishedVersion: cm.NewOptInt(3), severity: diag.SeverityError,
		},
		"higher response version": {
			responseVersion: 7, publishedVersion: cm.NewOptInt(3),
		},
		"nonpositive response version": {
			responseVersion: 0, publishedVersion: cm.NewOptInt(3), severity: diag.SeverityError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateEntryPublicationResponse(3, test.responseVersion, test.publishedVersion)
			if test.severity == 0 {
				require.Empty(t, diags)

				return
			}

			require.Len(t, diags, 1)
			assert.Equal(t, test.severity, diags[0].Severity())
		})
	}
}

func TestValidateEntryDraftResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		publishedVersion types.Int64
		hasError         bool
	}{
		"unpublished":         {publishedVersion: types.Int64Null()},
		"older publication":   {publishedVersion: types.Int64Value(1)},
		"recent publication":  {publishedVersion: types.Int64Value(2)},
		"same version":        {publishedVersion: types.Int64Value(3), hasError: true},
		"future publication":  {publishedVersion: types.Int64Value(4), hasError: true},
		"zero publication":    {publishedVersion: types.Int64Value(0)},
		"unknown publication": {publishedVersion: types.Int64Unknown(), hasError: true},
		"nonpositive version": {publishedVersion: types.Int64Null(), hasError: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			version := 3
			if name == "nonpositive version" {
				version = 0
			}

			assert.Equal(t, test.hasError, validateEntryDraftResponse(version, test.publishedVersion).HasError())
		})
	}
}
