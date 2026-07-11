package cmtesting_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
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
