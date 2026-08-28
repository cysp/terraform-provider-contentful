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
		normalPolicy     entryResponseFieldPolicy
		responseVersion  int
		publishedVersion cm.OptInt
		fieldPolicy      entryResponseFieldPolicy
		severity         diag.Severity
	}{
		"exact create response": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 4,
			publishedVersion: cm.NewOptInt(3), fieldPolicy: entryResponseFieldsCreationDefaults,
		},
		"exact update response": {
			normalPolicy: entryResponseFieldsExact, responseVersion: 4,
			publishedVersion: cm.NewOptInt(3), fieldPolicy: entryResponseFieldsExact,
		},
		"missing published version": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 4,
			fieldPolicy: entryResponseFieldsExact, severity: diag.SeverityError,
		},
		"wrong published version": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 4,
			publishedVersion: cm.NewOptInt(2), fieldPolicy: entryResponseFieldsExact, severity: diag.SeverityError,
		},
		"version did not advance": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 3,
			publishedVersion: cm.NewOptInt(3), fieldPolicy: entryResponseFieldsExact, severity: diag.SeverityWarning,
		},
		"higher response version": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 7,
			publishedVersion: cm.NewOptInt(3), fieldPolicy: entryResponseFieldsExact, severity: diag.SeverityWarning,
		},
		"nonpositive response version": {
			normalPolicy: entryResponseFieldsCreationDefaults, responseVersion: 0,
			publishedVersion: cm.NewOptInt(3), fieldPolicy: entryResponseFieldsExact, severity: diag.SeverityError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.fieldPolicy, entryPublicationResponseFieldPolicy(test.normalPolicy, 3, test.responseVersion, test.publishedVersion))

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
		"invalid publication": {publishedVersion: types.Int64Value(0), hasError: true},
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

func TestValidateObservedEntryLifecycle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version          int
		publishedVersion cm.OptInt
		hasError         bool
		hasWarning       bool
	}{
		"unpublished":           {version: 1},
		"published":             {version: 2, publishedVersion: cm.NewOptInt(1)},
		"pending draft":         {version: 4, publishedVersion: cm.NewOptInt(2)},
		"nonpositive version":   {version: 0, hasError: true},
		"nonpositive published": {version: 2, publishedVersion: cm.NewOptInt(0), hasError: true},
		"equal versions":        {version: 2, publishedVersion: cm.NewOptInt(2), hasWarning: true},
		"future published":      {version: 2, publishedVersion: cm.NewOptInt(3), hasWarning: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateObservedEntryLifecycle(test.version, test.publishedVersion)
			assert.Equal(t, test.hasError, diags.HasError())
			assert.Equal(t, test.hasWarning, diags.WarningsCount() > 0)
		})
	}
}

func TestValidateEntryStateLifecycle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		publishedVersion types.Int64
		hasError         bool
		hasWarning       bool
	}{
		"unpublished":         {publishedVersion: types.Int64Null()},
		"published":           {publishedVersion: types.Int64Value(1)},
		"pending draft":       {publishedVersion: types.Int64Value(2)},
		"unknown publication": {publishedVersion: types.Int64Unknown(), hasError: true},
		"nonpositive":         {publishedVersion: types.Int64Value(0), hasError: true},
		"equal version":       {publishedVersion: types.Int64Value(4), hasWarning: true},
		"future version":      {publishedVersion: types.Int64Value(5), hasWarning: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateEntryStateLifecycle(4, test.publishedVersion)
			assert.Equal(t, test.hasError, diags.HasError())
			assert.Equal(t, test.hasWarning, diags.WarningsCount() > 0)
		})
	}
}
