package cmtesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
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
	concept.URI, concept.PrefLabel = normalizeTaxonomyURI(req.URI), normalizeTaxonomyPrefLabel(req.PrefLabel)
	concept.AltLabels = normalizeTaxonomyLabels(concept.PrefLabel, req.AltLabels)
	concept.HiddenLabels = normalizeTaxonomyLabels(concept.PrefLabel, req.HiddenLabels)
	concept.Notations = make([]string, len(req.Notations))
	copy(concept.Notations, req.Notations)
	concept.Note = normalizeTaxonomyLocalizedString(req.Note)
	concept.ChangeNote = normalizeTaxonomyLocalizedString(req.ChangeNote)
	concept.Definition = normalizeTaxonomyLocalizedString(req.Definition)
	concept.EditorialNote = normalizeTaxonomyLocalizedString(req.EditorialNote)
	concept.Example = normalizeTaxonomyLocalizedString(req.Example)
	concept.HistoryNote = normalizeTaxonomyLocalizedString(req.HistoryNote)
	concept.ScopeNote = normalizeTaxonomyLocalizedString(req.ScopeNote)
	concept.Broader, concept.Related = req.Broader, req.Related
}

func normalizeTaxonomyPrefLabel(value cm.LocalizedString) cm.LocalizedString {
	result := cm.LocalizedString{}
	if label, ok := value["en-US"]; ok {
		result["en-US"] = label
	}

	return result
}

func normalizeTaxonomyURI(value cm.OptNilString) cm.OptNilString {
	if value.IsEmpty() {
		value.SetToNull()
	}

	return value
}

func normalizeTaxonomyLocalizedString(value cm.OptNilNullableLocalizedString) cm.OptNilNullableLocalizedString {
	if value.IsEmpty() {
		value.SetToNull()

		return value
	}

	localized, ok := value.Get()
	if !ok {
		return value
	}

	normalized := cm.NullableLocalizedString{}
	if label, ok := localized["en-US"]; ok {
		normalized["en-US"] = label
	}

	if len(normalized) == 0 {
		value.SetToNull()

		return value
	}

	value.SetTo(normalized)

	return value
}

func normalizeTaxonomyLabels(prefLabels cm.LocalizedString, labels cm.OptLocalizedStringList) cm.OptLocalizedStringList {
	configured, _ := labels.Get()

	normalized := cm.LocalizedStringList{}
	if labels, ok := configured["en-US"]; ok {
		normalized["en-US"] = labels
	}

	for locale := range prefLabels {
		if _, ok := normalized[locale]; !ok {
			normalized[locale] = []string{}
		}
	}

	return cm.NewOptLocalizedStringList(normalized)
}

type taxonomyLabelValidationDetails struct {
	Flatten struct {
		FormErrors  []string            `json:"formErrors"`
		FieldErrors map[string][]string `json:"fieldErrors"`
	} `json:"flatten"`
	Errors []taxonomyLabelValidationError `json:"errors"`
}

type taxonomyLabelValidationError struct {
	Name    string   `json:"name"`
	Path    []string `json:"path"`
	Details string   `json:"details"`
}

type taxonomyConceptLabelMaps struct {
	altLabels    cm.OptLocalizedStringList
	hiddenLabels cm.OptLocalizedStringList
}

func taxonomyConceptLabelValidationResponse(
	prefLabels cm.LocalizedString,
	labelMaps taxonomyConceptLabelMaps,
) (*cm.ErrorStatusCode, error) {
	const detail = "Invalid input: expected array, received undefined"

	details := taxonomyLabelValidationDetails{}
	details.Flatten.FormErrors = []string{}
	details.Flatten.FieldErrors = map[string][]string{}

	appendErrors := func(field string, labels cm.OptLocalizedStringList) {
		if !labels.IsSet() {
			return
		}

		configured, _ := labels.Get()

		locales := make([]string, 0, len(prefLabels))
		for locale := range prefLabels {
			locales = append(locales, locale)
		}

		slices.Sort(locales)

		for _, locale := range locales {
			if _, ok := configured[locale]; ok {
				continue
			}

			details.Flatten.FieldErrors[field] = append(details.Flatten.FieldErrors[field], detail)
			details.Errors = append(details.Errors, taxonomyLabelValidationError{
				Name: "invalid_type", Path: []string{field, locale}, Details: detail,
			})
		}
	}
	appendErrors("altLabels", labelMaps.altLabels)
	appendErrors("hiddenLabels", labelMaps.hiddenLabels)

	if len(details.Errors) == 0 {
		return nil, nil //nolint:nilnil // No validation response is the successful result.
	}

	encoded, err := json.Marshal(&details)
	if err != nil {
		return nil, fmt.Errorf("encode taxonomy label validation details: %w", err)
	}

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), encoded), nil
}

func taxonomyPrefLabelValidationResponse(prefLabels cm.LocalizedString) (*cm.ErrorStatusCode, error) {
	if _, ok := normalizeTaxonomyPrefLabel(prefLabels)["en-US"]; ok {
		return nil, nil //nolint:nilnil // No validation response is the successful result.
	}

	const detail = "Invalid input: expected string, received undefined"

	details := taxonomyLabelValidationDetails{}
	details.Flatten.FormErrors = []string{}
	details.Flatten.FieldErrors = map[string][]string{"prefLabel": {detail}}
	details.Errors = []taxonomyLabelValidationError{{Name: "invalid_type", Path: []string{"prefLabel", "en-US"}, Details: detail}}

	encoded, err := json.Marshal(&details)
	if err != nil {
		return nil, fmt.Errorf("encode taxonomy prefLabel validation details: %w", err)
	}

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), encoded), nil
}

func taxonomyLabelRemovalValidationResponse(field string) (*cm.ErrorStatusCode, error) {
	const detail = "Invalid input: expected object, received null"

	details := taxonomyLabelValidationDetails{}
	details.Flatten.FormErrors = []string{}
	details.Flatten.FieldErrors = map[string][]string{field: {detail}}
	details.Errors = []taxonomyLabelValidationError{{Name: "invalid_type", Path: []string{field}, Details: detail}}

	encoded, err := json.Marshal(&details)
	if err != nil {
		return nil, fmt.Errorf("encode taxonomy label removal validation details: %w", err)
	}

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), encoded), nil
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
	scheme.URI, scheme.PrefLabel = normalizeTaxonomyURI(req.URI), normalizeTaxonomyPrefLabel(req.PrefLabel)
	scheme.Definition = normalizeTaxonomyLocalizedString(req.Definition)
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
	switch request := current.(type) {
	case cm.TaxonomyConceptRequest:
		current = &request
	case cm.TaxonomyConceptSchemeRequest:
		current = &request
	}

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
		if !knownTaxonomyPatchField(current, key) {
			return fmt.Errorf("%w: %q", errUnknownTaxonomyPatchPath, operation.Path)
		}

		switch operation.Op {
		case cm.TaxonomyPatchItemOpAdd:
			fields[key] = json.RawMessage(operation.Value)
		case cm.TaxonomyPatchItemOpReplace, cm.TaxonomyPatchItemOpRemove:
			if _, ok := fields[key]; !ok {
				return fmt.Errorf("%w: %q", errUnknownTaxonomyPatchPath, operation.Path)
			}

			if operation.Op == cm.TaxonomyPatchItemOpReplace {
				fields[key] = json.RawMessage(operation.Value)

				continue
			}

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

func knownTaxonomyPatchField(current any, key string) bool {
	switch current.(type) {
	case *cm.TaxonomyConceptRequest:
		switch key {
		case "uri", "prefLabel", "altLabels", "hiddenLabels", "notations", "note", "changeNote", "definition", "editorialNote", "example", "historyNote", "scopeNote", "broader", "related":
			return true
		}
	case *cm.TaxonomyConceptSchemeRequest:
		switch key {
		case "uri", "prefLabel", "definition", "topConcepts", "concepts":
			return true
		}
	}

	return false
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

	prefLabelValidationResponse, err := taxonomyPrefLabelValidationResponse(req.PrefLabel)
	if err != nil {
		return nil, err
	}

	if prefLabelValidationResponse != nil {
		return prefLabelValidationResponse, nil
	}

	labelValidationResponse, err := taxonomyConceptLabelValidationResponse(normalizeTaxonomyPrefLabel(req.PrefLabel), taxonomyConceptLabelMaps{
		altLabels: req.AltLabels, hiddenLabels: req.HiddenLabels,
	})
	if err != nil {
		return nil, err
	}

	if labelValidationResponse != nil {
		return labelValidationResponse, nil
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

	if !updated.AltLabels.IsSet() {
		return taxonomyLabelRemovalValidationResponse("altLabels")
	}

	if !updated.HiddenLabels.IsSet() {
		return taxonomyLabelRemovalValidationResponse("hiddenLabels")
	}

	patchedLabelMaps := taxonomyConceptLabelMaps{}

	for _, operation := range req {
		if operation.Op != cm.TaxonomyPatchItemOpAdd && operation.Op != cm.TaxonomyPatchItemOpReplace {
			continue
		}

		switch operation.Path {
		case "/altLabels":
			patchedLabelMaps.altLabels = updated.AltLabels
		case "/hiddenLabels":
			patchedLabelMaps.hiddenLabels = updated.HiddenLabels
		}
	}

	prefLabelValidationResponse, err := taxonomyPrefLabelValidationResponse(updated.PrefLabel)
	if err != nil {
		return nil, err
	}

	if prefLabelValidationResponse != nil {
		return prefLabelValidationResponse, nil
	}

	labelValidationResponse, err := taxonomyConceptLabelValidationResponse(normalizeTaxonomyPrefLabel(updated.PrefLabel), patchedLabelMaps)
	if err != nil {
		return nil, err
	}

	if labelValidationResponse != nil {
		return labelValidationResponse, nil
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

	prefLabelValidationResponse, err := taxonomyPrefLabelValidationResponse(req.PrefLabel)
	if err != nil {
		return nil, err
	}

	if prefLabelValidationResponse != nil {
		return prefLabelValidationResponse, nil
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

	prefLabelValidationResponse, err := taxonomyPrefLabelValidationResponse(updated.PrefLabel)
	if err != nil {
		return nil, err
	}

	if prefLabelValidationResponse != nil {
		return prefLabelValidationResponse, nil
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
