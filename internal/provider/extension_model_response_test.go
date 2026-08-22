package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionResponsePreservesSourcePresence(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		response    cm.ExtensionExtension
		wantSrcNull bool
		wantSrcdoc  string
	}{
		"absent sources are null": {
			response:    cm.ExtensionExtension{Name: "Extension"},
			wantSrcNull: true,
		},
		"explicit empty srcdoc remains known empty": {
			response:    cm.ExtensionExtension{Name: "Extension", Srcdoc: cm.NewOptString("")},
			wantSrcNull: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := NewExtensionModelExtensionFromResponse(t.Context(), test.response)

			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, test.wantSrcNull, actual.Src.IsNull())

			if test.response.Srcdoc.IsSet() {
				assert.False(t, actual.SrcDoc.IsNull())
				assert.Equal(t, test.wantSrcdoc, actual.SrcDoc.ValueString())
			} else {
				assert.True(t, actual.SrcDoc.IsNull())
			}
		})
	}
}
