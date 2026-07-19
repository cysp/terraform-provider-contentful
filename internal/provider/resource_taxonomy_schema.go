package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxTaxonomyIDLength = 64

func taxonomyIDValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, maxTaxonomyIDLength),
		stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9._-]+$`), "must contain only letters, digits, dots, hyphens, or underscores"),
	}
}

func taxonomyIdentityAttributes(entityName string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{UseStateForUnknown()}},
		"organization_id": schema.StringAttribute{
			Description:   "ID of the organization that owns the " + entityName + ".",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
	}
}

func localizedStringAttribute(description string, required bool) schema.MapAttribute {
	return schema.MapAttribute{Description: description, Required: required, Optional: !required, ElementType: types.StringType}
}

func optionalComputedStringList(description string) schema.ListAttribute {
	return schema.ListAttribute{
		Description: description,
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		PlanModifiers: []planmodifier.List{
			UseStateForUnknown(),
		},
	}
}

func TaxonomyConceptResourceSchema(ctx context.Context) schema.Schema {
	attributes := taxonomyIdentityAttributes("taxonomy concept")
	attributes["concept_id"] = schema.StringAttribute{Description: "Caller-defined ID of the taxonomy concept.", Required: true, Validators: taxonomyIDValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}
	attributes["uri"] = schema.StringAttribute{Description: "Optional URI identifying the concept. Empty strings are rejected by Contentful.", Optional: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
	attributes["pref_label"] = localizedStringAttribute("Localized preferred labels.", true)
	attributes["alt_labels"] = schema.MapAttribute{
		Description: "Localized alternative labels.",
		Optional:    true,
		Computed:    true,
		ElementType: types.ListType{ElemType: types.StringType},
		PlanModifiers: []planmodifier.Map{
			UseStateForUnknown(),
		},
	}
	attributes["hidden_labels"] = schema.MapAttribute{
		Description: "Localized hidden labels.",
		Optional:    true,
		Computed:    true,
		ElementType: types.ListType{ElemType: types.StringType},
		PlanModifiers: []planmodifier.Map{
			UseStateForUnknown(),
		},
	}

	attributes["notations"] = optionalComputedStringList("Ordered notation values.")
	for name, description := range map[string]string{
		"note": "Localized notes.", "change_note": "Localized change notes.", "definition": "Localized definitions.",
		"editorial_note": "Localized editorial notes.", "example": "Localized examples.", "history_note": "Localized history notes.", "scope_note": "Localized scope notes.",
	} {
		attributes[name] = localizedStringAttribute(description, false)
	}

	attributes["broader_concept_ids"] = optionalComputedStringList("Ordered IDs of broader concepts.")
	attributes["related_concept_ids"] = optionalComputedStringList("Ordered IDs of related concepts.")
	attributes["concept_scheme_ids"] = schema.SetAttribute{Description: "IDs of schemes containing the concept.", Computed: true, ElementType: types.StringType}
	attributes["timeouts"] = timeouts.AttributesAll(ctx)

	return schema.Schema{Description: "Manages a Contentful taxonomy concept.", Attributes: attributes}
}

func TaxonomyConceptSchemeResourceSchema(ctx context.Context) schema.Schema {
	attributes := taxonomyIdentityAttributes("taxonomy concept scheme")
	attributes["concept_scheme_id"] = schema.StringAttribute{Description: "Caller-defined ID of the taxonomy concept scheme.", Required: true, Validators: taxonomyIDValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}
	attributes["uri"] = schema.StringAttribute{Description: "Optional URI identifying the concept scheme. Empty strings are rejected by Contentful.", Optional: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
	attributes["pref_label"] = localizedStringAttribute("Localized preferred labels.", true)
	attributes["definition"] = localizedStringAttribute("Localized definitions.", false)
	attributes["top_concept_ids"] = optionalComputedStringList("Ordered IDs of top concepts. Every top concept must also occur in concept_ids.")
	attributes["concept_ids"] = optionalComputedStringList("Ordered IDs of concepts in the scheme.")
	attributes["total_concepts"] = schema.Int64Attribute{Description: "Number of concepts in the scheme.", Computed: true}
	attributes["timeouts"] = timeouts.AttributesAll(ctx)

	return schema.Schema{Description: "Manages a Contentful taxonomy concept scheme.", Attributes: attributes}
}
