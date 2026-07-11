package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyHandlerLifecycleAndConstraints(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	handler := cmt.NewHandler()
	organizationID := "organization"

	parentRequest := cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Parent"}, Broader: []cm.TaxonomyConceptLink{}, Related: []cm.TaxonomyConceptLink{}}
	parentResponse, err := handler.PutTaxonomyConcept(ctx, &parentRequest, cm.PutTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "parent"})
	require.NoError(t, err)

	parent, ok := parentResponse.(*cm.TaxonomyConcept)
	require.True(t, ok)
	require.Equal(t, 1, parent.Sys.Version)
	altLabels, ok := parent.AltLabels.Get()
	require.True(t, ok)
	require.Equal(t, cm.LocalizedStringList{"en-US": {}}, altLabels)

	hiddenLabels, ok := parent.HiddenLabels.Get()
	require.True(t, ok)
	require.Equal(t, cm.LocalizedStringList{"en-US": {}}, hiddenLabels)

	selfRequest := cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Self"}, Broader: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("self")}}
	selfResponse, err := handler.PutTaxonomyConcept(ctx, &selfRequest, cm.PutTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "self"})
	require.NoError(t, err)

	selfStatus, ok := selfResponse.(cm.StatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusUnprocessableEntity, selfStatus.GetStatusCode())

	childRequest := cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Child"}, Broader: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent")}}
	childResponse, err := handler.PutTaxonomyConcept(ctx, &childRequest, cm.PutTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "child"})
	require.NoError(t, err)

	child, ok := childResponse.(*cm.TaxonomyConcept)
	require.True(t, ok)

	invalidSchemeRequest := cm.TaxonomyConceptSchemeRequest{PrefLabel: cm.LocalizedString{"en-US": "Invalid"}, TopConcepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent")}, Concepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("child")}}
	invalidSchemeResponse, err := handler.PutTaxonomyConceptScheme(ctx, &invalidSchemeRequest, cm.PutTaxonomyConceptSchemeParams{OrganizationID: organizationID, TaxonomyConceptSchemeID: "invalid"})
	require.NoError(t, err)

	invalidSchemeStatus, ok := invalidSchemeResponse.(cm.StatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusUnprocessableEntity, invalidSchemeStatus.GetStatusCode())

	schemeRequest := cm.TaxonomyConceptSchemeRequest{PrefLabel: cm.LocalizedString{"en-US": "Scheme"}, TopConcepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent")}, Concepts: []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("parent"), cm.NewTaxonomyConceptLink("child")}}
	_, err = handler.PutTaxonomyConceptScheme(ctx, &schemeRequest, cm.PutTaxonomyConceptSchemeParams{OrganizationID: organizationID, TaxonomyConceptSchemeID: "scheme"})
	require.NoError(t, err)
	require.Len(t, child.ConceptSchemes, 1)

	staleDelete, err := handler.DeleteTaxonomyConcept(ctx, cm.DeleteTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "parent", XContentfulVersion: 2})
	require.NoError(t, err)

	staleDeleteStatus, ok := staleDelete.(cm.StatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusConflict, staleDeleteStatus.GetStatusCode())

	deleted, err := handler.DeleteTaxonomyConcept(ctx, cm.DeleteTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "parent", XContentfulVersion: 1})
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deleted)
	require.Empty(t, child.Broader)

	childGetResponse, err := handler.GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{OrganizationID: organizationID, TaxonomyConceptID: "child"})
	require.NoError(t, err)
	require.Same(t, child, childGetResponse)
}

func TestTaxonomyConceptHandlerConstraints(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setup   []string
		id      string
		broader []string
		related []string
		status  int
	}{
		"missing broader concept": {id: "child", broader: []string{"missing"}, status: http.StatusUnprocessableEntity},
		"self related concept":    {id: "child", related: []string{"child"}, status: http.StatusUnprocessableEntity},
		"broader and related":     {setup: []string{"parent"}, id: "child", broader: []string{"parent"}, related: []string{"parent"}, status: http.StatusUnprocessableEntity},
		"missing related concept": {id: "child", related: []string{"missing"}, status: http.StatusUnprocessableEntity},
		"deduplicates links":      {setup: []string{"parent"}, id: "child", broader: []string{"parent", "parent"}, related: []string{}, status: http.StatusCreated},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := cmt.NewHandler()
			for _, conceptID := range test.setup {
				putTaxonomyConcept(t, handler, conceptID)
			}

			request := taxonomyConceptRequest(test.id, test.broader, test.related)
			response, err := handler.PutTaxonomyConcept(t.Context(), &request, cm.PutTaxonomyConceptParams{
				OrganizationID: "organization", TaxonomyConceptID: test.id,
			})
			require.NoError(t, err)

			if test.status == http.StatusCreated {
				concept, ok := response.(*cm.TaxonomyConcept)
				require.True(t, ok)
				require.Len(t, concept.Broader, 1)

				return
			}

			requireStatusCode(t, response, test.status)
		})
	}
}

func TestTaxonomyConceptSchemeHandlerConstraints(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	putTaxonomyConcept(t, handler, "concept")

	missingRequest := taxonomyConceptSchemeRequest([]string{"missing"}, nil)
	missingResponse, err := handler.PutTaxonomyConceptScheme(t.Context(), &missingRequest, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "missing",
	})
	require.NoError(t, err)
	requireStatusCode(t, missingResponse, http.StatusUnprocessableEntity)

	duplicateRequest := taxonomyConceptSchemeRequest([]string{"concept", "concept"}, []string{"concept"})
	response, err := handler.PutTaxonomyConceptScheme(t.Context(), &duplicateRequest, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)
	require.Len(t, scheme.Concepts, 1)
	require.Len(t, scheme.TopConcepts, 1)
	require.Equal(t, 1, scheme.TotalConcepts)

	duplicateResponse, err := handler.PutTaxonomyConceptScheme(t.Context(), &duplicateRequest, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)
	requireStatusCode(t, duplicateResponse, http.StatusConflict)
}

func TestTaxonomyConceptSchemeMembershipLifecycle(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	handler := cmt.NewHandler()
	putTaxonomyConcept(t, handler, "first")
	putTaxonomyConcept(t, handler, "second")
	putTaxonomyConcept(t, handler, "third")

	request := taxonomyConceptSchemeRequest([]string{"first", "second"}, []string{"first"})
	response, err := handler.PutTaxonomyConceptScheme(ctx, &request, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)
	require.IsType(t, &cm.TaxonomyConceptScheme{}, response)
	require.Len(t, getTaxonomyConcept(t, handler, "first").ConceptSchemes, 1)
	require.Len(t, getTaxonomyConcept(t, handler, "second").ConceptSchemes, 1)
	require.Empty(t, getTaxonomyConcept(t, handler, "third").ConceptSchemes)

	concepts := mustJSON(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("first"), cm.NewTaxonomyConceptLink("third")})
	topConcepts := mustJSON(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("first")})
	patchedResponse, err := handler.PatchTaxonomyConceptScheme(ctx, cm.TaxonomyPatch{
		{Op: cm.TaxonomyPatchItemOpAdd, Path: "/concepts", Value: jx.Raw(concepts)},
		{Op: cm.TaxonomyPatchItemOpAdd, Path: "/topConcepts", Value: jx.Raw(topConcepts)},
	}, cm.PatchTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme", XContentfulVersion: 1,
	})
	require.NoError(t, err)

	patched, ok := patchedResponse.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)
	require.Equal(t, 2, patched.Sys.Version)
	require.Len(t, getTaxonomyConcept(t, handler, "first").ConceptSchemes, 1)
	require.Empty(t, getTaxonomyConcept(t, handler, "second").ConceptSchemes)
	require.Len(t, getTaxonomyConcept(t, handler, "third").ConceptSchemes, 1)

	getResponse, err := handler.GetTaxonomyConceptScheme(ctx, cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)
	require.Same(t, patched, getResponse)

	deleted, err := handler.DeleteTaxonomyConceptScheme(ctx, cm.DeleteTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme", XContentfulVersion: 2,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deleted)
	require.Empty(t, getTaxonomyConcept(t, handler, "first").ConceptSchemes)
	require.Empty(t, getTaxonomyConcept(t, handler, "third").ConceptSchemes)
}

func TestTaxonomyConceptPatchLifecycle(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	handler := cmt.NewHandler()
	putTaxonomyConcept(t, handler, "related")
	concept := putTaxonomyConcept(t, handler, "concept")

	prefLabel := mustJSON(t, cm.LocalizedString{"en-US": "Updated concept"})
	related := mustJSON(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("related")})
	response, err := handler.PatchTaxonomyConcept(ctx, cm.TaxonomyPatch{
		{Op: cm.TaxonomyPatchItemOpAdd, Path: "/prefLabel", Value: jx.Raw(prefLabel)},
		{Op: cm.TaxonomyPatchItemOpAdd, Path: "/related", Value: jx.Raw(related)},
	}, cm.PatchTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: "concept", XContentfulVersion: 1,
	})
	require.NoError(t, err)

	updated, ok := response.(*cm.TaxonomyConcept)
	require.True(t, ok)
	require.Same(t, concept, updated)
	require.Equal(t, 2, updated.Sys.Version)
	require.Equal(t, "Updated concept", updated.PrefLabel["en-US"])
	require.Len(t, updated.Related, 1)

	staleResponse, err := handler.PatchTaxonomyConcept(ctx, nil, cm.PatchTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: "concept", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	requireStatusCode(t, staleResponse, http.StatusConflict)

	missingRelated := mustJSON(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("missing")})
	invalidResponse, err := handler.PatchTaxonomyConcept(ctx, cm.TaxonomyPatch{
		{Op: cm.TaxonomyPatchItemOpAdd, Path: "/related", Value: jx.Raw(missingRelated)},
	}, cm.PatchTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: "concept", XContentfulVersion: 2,
	})
	require.NoError(t, err)
	requireStatusCode(t, invalidResponse, http.StatusUnprocessableEntity)
	require.Equal(t, 2, concept.Sys.Version)
	require.Equal(t, "Updated concept", concept.PrefLabel["en-US"])
	require.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("related")}, concept.Related)
}

func TestTaxonomyPatchRejectsInvalidDocumentsAtomically(t *testing.T) {
	t.Parallel()

	tests := map[string]cm.TaxonomyPatch{
		"unknown path":      {{Op: cm.TaxonomyPatchItemOpAdd, Path: "/unknown", Value: jx.Raw(`null`)}},
		"nested path":       {{Op: cm.TaxonomyPatchItemOpAdd, Path: "/prefLabel/en-US", Value: jx.Raw(`"Changed"`)}},
		"missing slash":     {{Op: cm.TaxonomyPatchItemOpAdd, Path: "prefLabel", Value: jx.Raw(`{}`)}},
		"unknown operation": {{Op: cm.TaxonomyPatchItemOp("copy"), Path: "/prefLabel", Value: jx.Raw(`{}`)}},
	}

	for name, patch := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := cmt.NewHandler()
			concept := putTaxonomyConcept(t, handler, "concept")

			response, err := handler.PatchTaxonomyConcept(t.Context(), patch, cm.PatchTaxonomyConceptParams{
				OrganizationID: "organization", TaxonomyConceptID: "concept", XContentfulVersion: 1,
			})
			require.Error(t, err)
			require.Nil(t, response)

			persisted := getTaxonomyConcept(t, handler, "concept")
			require.Same(t, concept, persisted)
			require.Equal(t, 1, persisted.Sys.Version)
			require.Equal(t, "concept", persisted.PrefLabel["en-US"])
		})
	}
}

//nolint:paralleltest,tparallel
func TestTaxonomyConceptSchemePatchValidationIsAtomic(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	putTaxonomyConcept(t, handler, "member")
	putTaxonomyConcept(t, handler, "other")

	request := taxonomyConceptSchemeRequest([]string{"member"}, []string{"member"})
	response, err := handler.PutTaxonomyConceptScheme(t.Context(), &request, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)

	tests := map[string][]cm.TaxonomyConceptLink{
		"missing member":              {cm.NewTaxonomyConceptLink("missing")},
		"top concept outside members": {cm.NewTaxonomyConceptLink("other")},
	}

	for name, topConcepts := range tests {
		t.Run(name, func(t *testing.T) {
			patch := cm.TaxonomyPatch{}
			if name == "missing member" {
				patch = append(patch, cm.TaxonomyPatchItem{Op: cm.TaxonomyPatchItemOpAdd, Path: "/concepts", Value: jx.Raw(mustJSON(t, topConcepts))})
			} else {
				patch = append(patch, cm.TaxonomyPatchItem{Op: cm.TaxonomyPatchItemOpAdd, Path: "/topConcepts", Value: jx.Raw(mustJSON(t, topConcepts))})
			}

			patchResponse, err := handler.PatchTaxonomyConceptScheme(t.Context(), patch, cm.PatchTaxonomyConceptSchemeParams{
				OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme", XContentfulVersion: 1,
			})
			require.NoError(t, err)
			requireStatusCode(t, patchResponse, http.StatusUnprocessableEntity)

			persistedResponse, err := handler.GetTaxonomyConceptScheme(t.Context(), cm.GetTaxonomyConceptSchemeParams{
				OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
			})
			require.NoError(t, err)

			persisted, ok := persistedResponse.(*cm.TaxonomyConceptScheme)
			require.True(t, ok)
			require.Same(t, scheme, persisted)
			require.Equal(t, 1, persisted.Sys.Version)
			require.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("member")}, persisted.Concepts)
			require.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("member")}, persisted.TopConcepts)
			require.Len(t, getTaxonomyConcept(t, handler, "member").ConceptSchemes, 1)
			require.Empty(t, getTaxonomyConcept(t, handler, "other").ConceptSchemes)
		})
	}
}

func TestDeletingTaxonomyConceptUpdatesReferencesAndSchemes(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	putTaxonomyConcept(t, handler, "deleted")

	referencingRequest := taxonomyConceptRequest("referencing", []string{"deleted"}, nil)
	referencingResponse, err := handler.PutTaxonomyConcept(t.Context(), &referencingRequest, cm.PutTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: "referencing",
	})
	require.NoError(t, err)
	require.IsType(t, &cm.TaxonomyConcept{}, referencingResponse)

	schemeRequest := taxonomyConceptSchemeRequest([]string{"deleted", "referencing"}, []string{"deleted"})
	schemeResponse, err := handler.PutTaxonomyConceptScheme(t.Context(), &schemeRequest, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)
	require.IsType(t, &cm.TaxonomyConceptScheme{}, schemeResponse)

	deletedResponse, err := handler.DeleteTaxonomyConcept(t.Context(), cm.DeleteTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: "deleted", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.NoContent{}, deletedResponse)

	referencing := getTaxonomyConcept(t, handler, "referencing")
	require.Empty(t, referencing.Broader)
	require.Len(t, referencing.ConceptSchemes, 1)
	require.Equal(t, "scheme", referencing.ConceptSchemes[0].Sys.ID)

	updatedSchemeResponse, err := handler.GetTaxonomyConceptScheme(t.Context(), cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)

	updatedScheme, ok := updatedSchemeResponse.(*cm.TaxonomyConceptScheme)
	require.True(t, ok)
	require.Equal(t, []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("referencing")}, updatedScheme.Concepts)
	require.Empty(t, updatedScheme.TopConcepts)
	require.Equal(t, 1, updatedScheme.TotalConcepts)
}

func TestTaxonomyHandlerNotFoundAndVersionConflicts(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	handler := cmt.NewHandler()

	conceptResponses := make([]any, 0, 3)
	getConceptResponse, err := handler.GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{OrganizationID: "organization", TaxonomyConceptID: "missing"})
	require.NoError(t, err)

	conceptResponses = append(conceptResponses, getConceptResponse)
	patchConceptResponse, err := handler.PatchTaxonomyConcept(ctx, nil, cm.PatchTaxonomyConceptParams{OrganizationID: "organization", TaxonomyConceptID: "missing", XContentfulVersion: 1})
	require.NoError(t, err)

	conceptResponses = append(conceptResponses, patchConceptResponse)
	deleteConceptResponse, err := handler.DeleteTaxonomyConcept(ctx, cm.DeleteTaxonomyConceptParams{OrganizationID: "organization", TaxonomyConceptID: "missing", XContentfulVersion: 1})
	require.NoError(t, err)

	conceptResponses = append(conceptResponses, deleteConceptResponse)

	for _, missing := range conceptResponses {
		requireStatusCode(t, missing, http.StatusNotFound)
	}

	schemeResponses := make([]any, 0, 3)
	getSchemeResponse, err := handler.GetTaxonomyConceptScheme(ctx, cm.GetTaxonomyConceptSchemeParams{OrganizationID: "organization", TaxonomyConceptSchemeID: "missing"})
	require.NoError(t, err)

	schemeResponses = append(schemeResponses, getSchemeResponse)
	patchSchemeResponse, err := handler.PatchTaxonomyConceptScheme(ctx, nil, cm.PatchTaxonomyConceptSchemeParams{OrganizationID: "organization", TaxonomyConceptSchemeID: "missing", XContentfulVersion: 1})
	require.NoError(t, err)

	schemeResponses = append(schemeResponses, patchSchemeResponse)
	deleteSchemeResponse, err := handler.DeleteTaxonomyConceptScheme(ctx, cm.DeleteTaxonomyConceptSchemeParams{OrganizationID: "organization", TaxonomyConceptSchemeID: "missing", XContentfulVersion: 1})
	require.NoError(t, err)

	schemeResponses = append(schemeResponses, deleteSchemeResponse)

	for _, missing := range schemeResponses {
		requireStatusCode(t, missing, http.StatusNotFound)
	}

	putTaxonomyConcept(t, handler, "concept")

	schemeRequest := taxonomyConceptSchemeRequest([]string{"concept"}, []string{"concept"})
	_, err = handler.PutTaxonomyConceptScheme(ctx, &schemeRequest, cm.PutTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme",
	})
	require.NoError(t, err)

	stalePatch, err := handler.PatchTaxonomyConceptScheme(ctx, nil, cm.PatchTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme", XContentfulVersion: 2,
	})
	require.NoError(t, err)
	requireStatusCode(t, stalePatch, http.StatusConflict)

	staleDelete, err := handler.DeleteTaxonomyConceptScheme(ctx, cm.DeleteTaxonomyConceptSchemeParams{
		OrganizationID: "organization", TaxonomyConceptSchemeID: "scheme", XContentfulVersion: 2,
	})
	require.NoError(t, err)
	requireStatusCode(t, staleDelete, http.StatusConflict)
}

func putTaxonomyConcept(t *testing.T, handler *cmt.Handler, conceptID string) *cm.TaxonomyConcept {
	t.Helper()

	request := taxonomyConceptRequest(conceptID, nil, nil)
	response, err := handler.PutTaxonomyConcept(t.Context(), &request, cm.PutTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: conceptID,
	})
	require.NoError(t, err)

	concept, ok := response.(*cm.TaxonomyConcept)
	require.True(t, ok)

	return concept
}

func getTaxonomyConcept(t *testing.T, handler *cmt.Handler, conceptID string) *cm.TaxonomyConcept {
	t.Helper()

	response, err := handler.GetTaxonomyConcept(t.Context(), cm.GetTaxonomyConceptParams{
		OrganizationID: "organization", TaxonomyConceptID: conceptID,
	})
	require.NoError(t, err)

	concept, ok := response.(*cm.TaxonomyConcept)
	require.True(t, ok)

	return concept
}

func taxonomyConceptRequest(conceptID string, broader, related []string) cm.TaxonomyConceptRequest {
	var localizedNull cm.OptNilNullableLocalizedString
	localizedNull.SetToNull()

	return cm.TaxonomyConceptRequest{
		URI:           cm.NewOptNilPointerString(nil),
		PrefLabel:     cm.LocalizedString{"en-US": conceptID},
		AltLabels:     cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		HiddenLabels:  cm.NewOptLocalizedStringList(cm.LocalizedStringList{}),
		Notations:     []string{},
		Note:          localizedNull,
		ChangeNote:    localizedNull,
		Definition:    localizedNull,
		EditorialNote: localizedNull,
		Example:       localizedNull,
		HistoryNote:   localizedNull,
		ScopeNote:     localizedNull,
		Broader:       conceptLinks(broader),
		Related:       conceptLinks(related),
	}
}

func taxonomyConceptSchemeRequest(concepts, topConcepts []string) cm.TaxonomyConceptSchemeRequest {
	var definition cm.OptNilNullableLocalizedString
	definition.SetToNull()

	return cm.TaxonomyConceptSchemeRequest{
		URI:         cm.NewOptNilPointerString(nil),
		PrefLabel:   cm.LocalizedString{"en-US": "Scheme"},
		Definition:  definition,
		Concepts:    conceptLinks(concepts),
		TopConcepts: conceptLinks(topConcepts),
	}
}

func conceptLinks(ids []string) []cm.TaxonomyConceptLink {
	links := make([]cm.TaxonomyConceptLink, 0, len(ids))
	for _, id := range ids {
		links = append(links, cm.NewTaxonomyConceptLink(id))
	}

	return links
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return data
}

func requireStatusCode(t *testing.T, response any, expected int) {
	t.Helper()

	status, ok := response.(cm.StatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, expected, status.GetStatusCode())
}
