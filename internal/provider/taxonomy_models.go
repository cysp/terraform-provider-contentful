package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TaxonomyConceptIdentityModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	ConceptID      types.String `tfsdk:"concept_id"`
}

type TaxonomyConceptModel struct {
	IDIdentityModel
	TaxonomyConceptIdentityModel

	URI               types.String `tfsdk:"uri"`
	PrefLabel         types.Map    `tfsdk:"pref_label"`
	AltLabels         types.Map    `tfsdk:"alt_labels"`
	HiddenLabels      types.Map    `tfsdk:"hidden_labels"`
	Notations         types.List   `tfsdk:"notations"`
	Note              types.Map    `tfsdk:"note"`
	ChangeNote        types.Map    `tfsdk:"change_note"`
	Definition        types.Map    `tfsdk:"definition"`
	EditorialNote     types.Map    `tfsdk:"editorial_note"`
	Example           types.Map    `tfsdk:"example"`
	HistoryNote       types.Map    `tfsdk:"history_note"`
	ScopeNote         types.Map    `tfsdk:"scope_note"`
	BroaderConceptIDs types.List   `tfsdk:"broader_concept_ids"`
	RelatedConceptIDs types.List   `tfsdk:"related_concept_ids"`
	ConceptSchemeIDs  types.Set    `tfsdk:"concept_scheme_ids"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type TaxonomyConceptSchemeIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	ConceptSchemeID types.String `tfsdk:"concept_scheme_id"`
}

type TaxonomyConceptSchemeModel struct {
	IDIdentityModel
	TaxonomyConceptSchemeIdentityModel

	URI           types.String `tfsdk:"uri"`
	PrefLabel     types.Map    `tfsdk:"pref_label"`
	Definition    types.Map    `tfsdk:"definition"`
	TopConceptIDs types.List   `tfsdk:"top_concept_ids"`
	ConceptIDs    types.List   `tfsdk:"concept_ids"`
	TotalConcepts types.Int64  `tfsdk:"total_concepts"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func stringMap(ctx context.Context, value types.Map) (map[string]string, diag.Diagnostics) {
	result := map[string]string{}
	if value.IsNull() || value.IsUnknown() {
		return result, nil
	}

	diags := value.ElementsAs(ctx, &result, false)

	return result, diags
}

func stringList(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	result := []string{}
	if value.IsNull() || value.IsUnknown() {
		return result, nil
	}

	diags := value.ElementsAs(ctx, &result, false)

	return result, diags
}

func stringListMap(ctx context.Context, value types.Map) (map[string][]string, diag.Diagnostics) {
	result := map[string][]string{}
	if value.IsNull() || value.IsUnknown() {
		return result, nil
	}

	diags := value.ElementsAs(ctx, &result, false)

	return result, diags
}
