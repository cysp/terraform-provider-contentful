package provider

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/require"
)

func completeTaxonomyConceptRequest() cm.TaxonomyConceptRequest {
	return cm.TaxonomyConceptRequest{
		URI:           cm.NewOptNilPointerString(new("https://example.com/concepts/chair")),
		PrefLabel:     cm.LocalizedString{"en-US": "Chair"},
		AltLabels:     cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {"Seat"}}),
		HiddenLabels:  cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		Notations:     []string{"CHAIR"},
		Note:          nullLocalizedString(),
		ChangeNote:    nullLocalizedString(),
		Definition:    cm.NewOptNilNullableLocalizedString(cm.NullableLocalizedString{"en-US": "A seat"}),
		EditorialNote: nullLocalizedString(),
		Example:       nullLocalizedString(),
		HistoryNote:   nullLocalizedString(),
		ScopeNote:     nullLocalizedString(),
		Broader:       []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("furniture")},
		Related:       []cm.TaxonomyConceptLink{},
	}
}

func cloneTaxonomyConceptRequest(t *testing.T, request cm.TaxonomyConceptRequest) cm.TaxonomyConceptRequest {
	t.Helper()

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var result cm.TaxonomyConceptRequest
	require.NoError(t, json.Unmarshal(data, &result))

	return result
}

func nullLocalizedString() cm.OptNilNullableLocalizedString {
	var value cm.OptNilNullableLocalizedString
	value.SetToNull()

	return value
}

func assertTaxonomyDiagnosticPath(t *testing.T, diags diag.Diagnostics, want path.Path) {
	t.Helper()

	for _, diagnostic := range diags.Errors() {
		pathDiagnostic, ok := diagnostic.(diag.DiagnosticWithPath)
		if ok && pathDiagnostic.Path().String() == want.String() {
			return
		}
	}

	t.Fatalf("diagnostics did not include path %s: %v", want, diags)
}
