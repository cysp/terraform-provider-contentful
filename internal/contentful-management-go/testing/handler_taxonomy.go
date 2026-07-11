package cmtesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

var (
	errInvalidTaxonomyPatchPath   = errors.New("invalid taxonomy patch path")
	errUnknownTaxonomyPatchPath   = errors.New("unknown taxonomy patch path")
	errUnsupportedTaxonomyPatchOp = errors.New("unsupported taxonomy patch operation")
)

func taxonomyConceptFromRequest(organizationID, id string, req *cm.TaxonomyConceptRequest) cm.TaxonomyConcept {
	concept := cm.TaxonomyConcept{
		Sys:            cm.TaxonomyConceptSys{Organization: cm.NewOrganizationLink(organizationID), Type: cm.TaxonomyConceptSysTypeTaxonomyConcept, ID: id},
		ConceptSchemes: []cm.TaxonomyConceptSchemeLink{},
	}
	updateTaxonomyConcept(&concept, req)

	return concept
}

func updateTaxonomyConcept(concept *cm.TaxonomyConcept, req *cm.TaxonomyConceptRequest) {
	concept.Sys.Version++
	concept.URI, concept.PrefLabel = req.URI, req.PrefLabel
	concept.AltLabels = normalizeTaxonomyLabels(req.PrefLabel, req.AltLabels)
	concept.HiddenLabels = normalizeTaxonomyLabels(req.PrefLabel, req.HiddenLabels)
	concept.Notations, concept.Note, concept.ChangeNote, concept.Definition = req.Notations, req.Note, req.ChangeNote, req.Definition
	concept.EditorialNote, concept.Example, concept.HistoryNote, concept.ScopeNote = req.EditorialNote, req.Example, req.HistoryNote, req.ScopeNote
	concept.Broader, concept.Related = req.Broader, req.Related
}

func normalizeTaxonomyLabels(prefLabels cm.LocalizedString, labels cm.OptLocalizedStringList) cm.OptLocalizedStringList {
	configured, _ := labels.Get()
	normalized := make(cm.LocalizedStringList, len(configured))
	maps.Copy(normalized, configured)

	for locale := range prefLabels {
		if _, ok := normalized[locale]; !ok {
			normalized[locale] = []string{}
		}
	}

	return cm.NewOptLocalizedStringList(normalized)
}

func taxonomyConceptSchemeFromRequest(organizationID, id string, req *cm.TaxonomyConceptSchemeRequest) cm.TaxonomyConceptScheme {
	scheme := cm.TaxonomyConceptScheme{Sys: cm.TaxonomyConceptSchemeSys{
		Organization: cm.NewOrganizationLink(organizationID), Type: cm.TaxonomyConceptSchemeSysTypeTaxonomyConceptScheme, ID: id,
	}}
	updateTaxonomyConceptScheme(&scheme, req)

	return scheme
}

func updateTaxonomyConceptScheme(scheme *cm.TaxonomyConceptScheme, req *cm.TaxonomyConceptSchemeRequest) {
	scheme.Sys.Version++
	scheme.URI, scheme.PrefLabel, scheme.Definition = req.URI, req.PrefLabel, req.Definition
	scheme.TopConcepts, scheme.Concepts, scheme.TotalConcepts = req.TopConcepts, req.Concepts, len(req.Concepts)
}

func (h *Handler) validateConceptLinks(organizationID, conceptID string, req *cm.TaxonomyConceptRequest) string {
	broader := map[string]bool{}

	for _, link := range req.Broader {
		if link.Sys.ID == conceptID {
			return "Concept can't reference itself"
		}

		if h.taxonomyConcepts.Get(organizationID, link.Sys.ID) == nil {
			return "Failed to find concept: " + link.Sys.ID
		}

		broader[link.Sys.ID] = true
	}

	for _, link := range req.Related {
		if link.Sys.ID == conceptID {
			return "Concept can't reference itself"
		}

		if broader[link.Sys.ID] {
			return "Concept can't be related and broader at the same time."
		}

		if h.taxonomyConcepts.Get(organizationID, link.Sys.ID) == nil {
			return "Failed to find concept: " + link.Sys.ID
		}
	}

	return ""
}

func deduplicateConceptLinks(links []cm.TaxonomyConceptLink) []cm.TaxonomyConceptLink {
	seen := map[string]bool{}

	result := make([]cm.TaxonomyConceptLink, 0, len(links))
	for _, link := range links {
		if seen[link.Sys.ID] {
			continue
		}

		seen[link.Sys.ID] = true
		result = append(result, cm.NewTaxonomyConceptLink(link.Sys.ID))
	}

	return result
}

func (h *Handler) validateSchemeLinks(organizationID string, req *cm.TaxonomyConceptSchemeRequest) string {
	members := map[string]bool{}

	for _, link := range req.Concepts {
		if h.taxonomyConcepts.Get(organizationID, link.Sys.ID) == nil {
			return "Failed to find concept: " + link.Sys.ID
		}

		members[link.Sys.ID] = true
	}

	for _, link := range req.TopConcepts {
		if !members[link.Sys.ID] {
			return "Top concepts must be in scheme."
		}
	}

	return ""
}

func applyTaxonomyPatch(current any, patch cm.TaxonomyPatch, destination any) error {
	data, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("encode current taxonomy request: %w", err)
	}

	fields := map[string]json.RawMessage{}

	err = json.Unmarshal(data, &fields)
	if err != nil {
		return fmt.Errorf("decode current taxonomy request: %w", err)
	}

	for _, operation := range patch {
		if !strings.HasPrefix(operation.Path, "/") || strings.Contains(strings.TrimPrefix(operation.Path, "/"), "/") {
			return fmt.Errorf("%w: %q", errInvalidTaxonomyPatchPath, operation.Path)
		}

		key := strings.TrimPrefix(operation.Path, "/")
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("%w: %q", errUnknownTaxonomyPatchPath, operation.Path)
		}

		switch operation.Op {
		case cm.TaxonomyPatchItemOpAdd, cm.TaxonomyPatchItemOpReplace:
			fields[key] = json.RawMessage(operation.Value)
		case cm.TaxonomyPatchItemOpRemove:
			delete(fields, key)
		default:
			return fmt.Errorf("%w: %q", errUnsupportedTaxonomyPatchOp, operation.Op)
		}
	}

	data, err = json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("encode patched taxonomy request: %w", err)
	}

	err = json.Unmarshal(data, destination)
	if err != nil {
		return fmt.Errorf("decode patched taxonomy request: %w", err)
	}

	return nil
}

//nolint:ireturn
func (h *Handler) GetTaxonomyConcept(_ context.Context, params cm.GetTaxonomyConceptParams) (cm.GetTaxonomyConceptRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	concept := h.taxonomyConcepts.Get(params.OrganizationID, params.TaxonomyConceptID)
	if concept == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept not found"), nil), nil
	}

	return concept, nil
}

//nolint:ireturn
func (h *Handler) PutTaxonomyConcept(_ context.Context, req *cm.TaxonomyConceptRequest, params cm.PutTaxonomyConceptParams) (cm.PutTaxonomyConceptRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.taxonomyConcepts.Get(params.OrganizationID, params.TaxonomyConceptID) != nil {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	validationMessage := h.validateConceptLinks(params.OrganizationID, params.TaxonomyConceptID, req)
	if validationMessage != "" {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new(validationMessage), nil), nil
	}

	req.Broader = deduplicateConceptLinks(req.Broader)
	req.Related = deduplicateConceptLinks(req.Related)
	concept := taxonomyConceptFromRequest(params.OrganizationID, params.TaxonomyConceptID, req)
	h.taxonomyConcepts.Set(params.OrganizationID, params.TaxonomyConceptID, &concept)

	return &concept, nil
}

//nolint:ireturn
func (h *Handler) PatchTaxonomyConcept(_ context.Context, req cm.TaxonomyPatch, params cm.PatchTaxonomyConceptParams) (cm.PatchTaxonomyConceptRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	concept := h.taxonomyConcepts.Get(params.OrganizationID, params.TaxonomyConceptID)
	if concept == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept not found"), nil), nil
	}

	if params.XContentfulVersion != concept.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	current := cm.TaxonomyConceptRequest{URI: concept.URI, PrefLabel: concept.PrefLabel, AltLabels: concept.AltLabels, HiddenLabels: concept.HiddenLabels, Notations: concept.Notations, Note: concept.Note, ChangeNote: concept.ChangeNote, Definition: concept.Definition, EditorialNote: concept.EditorialNote, Example: concept.Example, HistoryNote: concept.HistoryNote, ScopeNote: concept.ScopeNote, Broader: concept.Broader, Related: concept.Related}

	var updated cm.TaxonomyConceptRequest

	err := applyTaxonomyPatch(current, req, &updated)
	if err != nil {
		return nil, err
	}

	validationMessage := h.validateConceptLinks(params.OrganizationID, params.TaxonomyConceptID, &updated)
	if validationMessage != "" {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new(validationMessage), nil), nil
	}

	updated.Broader = deduplicateConceptLinks(updated.Broader)
	updated.Related = deduplicateConceptLinks(updated.Related)
	updateTaxonomyConcept(concept, &updated)

	return concept, nil
}

//nolint:ireturn
func (h *Handler) DeleteTaxonomyConcept(_ context.Context, params cm.DeleteTaxonomyConceptParams) (cm.DeleteTaxonomyConceptRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	concept := h.taxonomyConcepts.Get(params.OrganizationID, params.TaxonomyConceptID)
	if concept == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept not found"), nil), nil
	}

	if params.XContentfulVersion != concept.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	h.taxonomyConcepts.Delete(params.OrganizationID, params.TaxonomyConceptID)

	for _, other := range h.taxonomyConcepts.Values(params.OrganizationID) {
		other.Broader = linksWithoutID(other.Broader, params.TaxonomyConceptID)
		other.Related = linksWithoutID(other.Related, params.TaxonomyConceptID)
	}

	for _, scheme := range h.taxonomyConceptSchemes.Values(params.OrganizationID) {
		scheme.TopConcepts = linksWithoutID(scheme.TopConcepts, params.TaxonomyConceptID)
		scheme.Concepts = linksWithoutID(scheme.Concepts, params.TaxonomyConceptID)
		scheme.TotalConcepts = len(scheme.Concepts)
	}

	return &cm.NoContent{}, nil
}

func linksWithoutID(links []cm.TaxonomyConceptLink, id string) []cm.TaxonomyConceptLink {
	result := links[:0]
	for _, link := range links {
		if link.Sys.ID != id {
			result = append(result, link)
		}
	}

	return result
}

//nolint:ireturn
func (h *Handler) GetTaxonomyConceptScheme(_ context.Context, params cm.GetTaxonomyConceptSchemeParams) (cm.GetTaxonomyConceptSchemeRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	scheme := h.taxonomyConceptSchemes.Get(params.OrganizationID, params.TaxonomyConceptSchemeID)
	if scheme == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept scheme not found"), nil), nil
	}

	return scheme, nil
}

//nolint:ireturn
func (h *Handler) PutTaxonomyConceptScheme(_ context.Context, req *cm.TaxonomyConceptSchemeRequest, params cm.PutTaxonomyConceptSchemeParams) (cm.PutTaxonomyConceptSchemeRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.taxonomyConceptSchemes.Get(params.OrganizationID, params.TaxonomyConceptSchemeID) != nil {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	validationMessage := h.validateSchemeLinks(params.OrganizationID, req)
	if validationMessage != "" {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new(validationMessage), nil), nil
	}

	req.TopConcepts = deduplicateConceptLinks(req.TopConcepts)
	req.Concepts = deduplicateConceptLinks(req.Concepts)
	scheme := taxonomyConceptSchemeFromRequest(params.OrganizationID, params.TaxonomyConceptSchemeID, req)
	h.taxonomyConceptSchemes.Set(params.OrganizationID, params.TaxonomyConceptSchemeID, &scheme)
	h.syncConceptSchemeMembership(params.OrganizationID, params.TaxonomyConceptSchemeID, nil, scheme.Concepts)

	return &scheme, nil
}

//nolint:ireturn
func (h *Handler) PatchTaxonomyConceptScheme(_ context.Context, req cm.TaxonomyPatch, params cm.PatchTaxonomyConceptSchemeParams) (cm.PatchTaxonomyConceptSchemeRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	scheme := h.taxonomyConceptSchemes.Get(params.OrganizationID, params.TaxonomyConceptSchemeID)
	if scheme == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept scheme not found"), nil), nil
	}

	if params.XContentfulVersion != scheme.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	oldConcepts := append([]cm.TaxonomyConceptLink(nil), scheme.Concepts...)
	current := cm.TaxonomyConceptSchemeRequest{URI: scheme.URI, PrefLabel: scheme.PrefLabel, Definition: scheme.Definition, TopConcepts: scheme.TopConcepts, Concepts: scheme.Concepts}

	var updated cm.TaxonomyConceptSchemeRequest

	err := applyTaxonomyPatch(current, req, &updated)
	if err != nil {
		return nil, err
	}

	validationMessage := h.validateSchemeLinks(params.OrganizationID, &updated)
	if validationMessage != "" {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new(validationMessage), nil), nil
	}

	updated.TopConcepts = deduplicateConceptLinks(updated.TopConcepts)
	updated.Concepts = deduplicateConceptLinks(updated.Concepts)
	updateTaxonomyConceptScheme(scheme, &updated)
	h.syncConceptSchemeMembership(params.OrganizationID, params.TaxonomyConceptSchemeID, oldConcepts, scheme.Concepts)

	return scheme, nil
}

//nolint:ireturn
func (h *Handler) DeleteTaxonomyConceptScheme(_ context.Context, params cm.DeleteTaxonomyConceptSchemeParams) (cm.DeleteTaxonomyConceptSchemeRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	scheme := h.taxonomyConceptSchemes.Get(params.OrganizationID, params.TaxonomyConceptSchemeID)
	if scheme == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Taxonomy concept scheme not found"), nil), nil
	}

	if params.XContentfulVersion != scheme.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	h.syncConceptSchemeMembership(params.OrganizationID, params.TaxonomyConceptSchemeID, scheme.Concepts, nil)
	h.taxonomyConceptSchemes.Delete(params.OrganizationID, params.TaxonomyConceptSchemeID)

	return &cm.NoContent{}, nil
}

func (h *Handler) syncConceptSchemeMembership(organizationID, schemeID string, oldLinks, newLinks []cm.TaxonomyConceptLink) {
	oldIDs, newIDs := map[string]bool{}, map[string]bool{}
	for _, link := range oldLinks {
		oldIDs[link.Sys.ID] = true
	}

	for _, link := range newLinks {
		newIDs[link.Sys.ID] = true
	}

	for id := range oldIDs {
		if newIDs[id] {
			continue
		}

		if concept := h.taxonomyConcepts.Get(organizationID, id); concept != nil {
			filtered := concept.ConceptSchemes[:0]
			for _, link := range concept.ConceptSchemes {
				if link.Sys.ID != schemeID {
					filtered = append(filtered, link)
				}
			}

			concept.ConceptSchemes = filtered
		}
	}

	for id := range newIDs {
		if oldIDs[id] {
			continue
		}

		if concept := h.taxonomyConcepts.Get(organizationID, id); concept != nil {
			concept.ConceptSchemes = append(concept.ConceptSchemes, newTaxonomyConceptSchemeLink(schemeID))
		}
	}
}

func newTaxonomyConceptSchemeLink(id string) cm.TaxonomyConceptSchemeLink {
	return cm.TaxonomyConceptSchemeLink{Sys: cm.TaxonomyConceptSchemeLinkSys{
		Type: cm.TaxonomyConceptSchemeLinkSysTypeLink, ID: id, LinkType: cm.TaxonomyConceptSchemeLinkSysLinkTypeTaxonomyConceptScheme,
	}}
}
