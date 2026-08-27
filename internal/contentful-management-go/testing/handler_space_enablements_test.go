//nolint:testpackage // The absent-document assertion deliberately verifies the same-package fake store.
package cmtesting

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutSpaceEnablementsRejectsInvalidPairedMembersWithoutMutation(t *testing.T) {
	t.Parallel()

	tests := map[string]cm.SpaceEnablementData{
		"both omitted": {},
		"cross-space links only": {
			CrossSpaceLinks: enabledSpaceEnablementField(true),
		},
		"space templates only": {
			SpaceTemplates: enabledSpaceEnablementField(false),
		},
		"unequal values": {
			CrossSpaceLinks: enabledSpaceEnablementField(true),
			SpaceTemplates:  enabledSpaceEnablementField(false),
		},
	}

	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler, before := newSpaceEnablementsTestHandler(t)

			response, err := handler.PutSpaceEnablements(t.Context(), &request, cm.PutSpaceEnablementsParams{
				SpaceID:            "space",
				XContentfulVersion: before.Sys.Version,
			})
			require.NoError(t, err)
			requireSpaceEnablementsError(t, response, http.StatusUnprocessableEntity, "ValidationFailed")
			assert.Equal(t, before, *handler.enablements["space"])
		})
	}
}

func TestPutSpaceEnablementsChecksSpaceBeforePairedMembers(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	orphaned := NewSpaceEnablementFromRequestFields("missing", validSpaceEnablementData(false))
	handler.enablements["missing"] = &orphaned
	before := orphaned
	request := cm.SpaceEnablementData{}

	response, err := handler.PutSpaceEnablements(t.Context(), &request, cm.PutSpaceEnablementsParams{
		SpaceID:            "missing",
		XContentfulVersion: before.Sys.Version,
	})
	require.NoError(t, err)
	requireSpaceEnablementsError(t, response, http.StatusNotFound, "NotFound")
	assert.Equal(t, before, *handler.enablements["missing"])
}

func TestPutSpaceEnablementsChecksVersionBeforePairedMembers(t *testing.T) {
	t.Parallel()

	handler, before := newSpaceEnablementsTestHandler(t)
	request := cm.SpaceEnablementData{}

	response, err := handler.PutSpaceEnablements(t.Context(), &request, cm.PutSpaceEnablementsParams{
		SpaceID:            "space",
		XContentfulVersion: before.Sys.Version + 1,
	})
	require.NoError(t, err)
	requireSpaceEnablementsError(t, response, http.StatusConflict, "VersionMismatch")
	assert.Equal(t, before, *handler.enablements["space"])
}

func TestPutSpaceEnablementsDoesNotCreateDocumentForInvalidPairedMembers(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")

	request := cm.SpaceEnablementData{}

	response, err := handler.PutSpaceEnablements(t.Context(), &request, cm.PutSpaceEnablementsParams{
		SpaceID:            "space",
		XContentfulVersion: 1,
	})
	require.NoError(t, err)
	requireSpaceEnablementsError(t, response, http.StatusUnprocessableEntity, "ValidationFailed")

	_, stored := handler.enablements["space"]
	assert.False(t, stored)
}

func TestPutSpaceEnablementsStoresValidPairedMembersAndAdvancesVersionOnce(t *testing.T) {
	t.Parallel()

	handler, before := newSpaceEnablementsTestHandler(t)

	falseRequest := validSpaceEnablementData(false)
	falseRequest.StudioExperiences = enabledSpaceEnablementField(false)
	falseRequest.SuggestConcepts = enabledSpaceEnablementField(true)
	falseResponse, err := handler.PutSpaceEnablements(t.Context(), &falseRequest, cm.PutSpaceEnablementsParams{
		SpaceID:            "space",
		XContentfulVersion: before.Sys.Version,
	})
	require.NoError(t, err)
	storedFalse := requireSpaceEnablementsSuccess(t, falseResponse)
	assert.Equal(t, before.Sys.Version+1, storedFalse.Sys.Version)
	assert.Equal(t, falseRequest.CrossSpaceLinks, storedFalse.CrossSpaceLinks)
	assert.Equal(t, falseRequest.SpaceTemplates, storedFalse.SpaceTemplates)
	assert.Equal(t, falseRequest.StudioExperiences, storedFalse.StudioExperiences)
	assert.Equal(t, falseRequest.SuggestConcepts, storedFalse.SuggestConcepts)
	assert.Equal(t, storedFalse, *handler.enablements["space"])

	trueRequest := validSpaceEnablementData(true)
	trueRequest.StudioExperiences = enabledSpaceEnablementField(true)
	trueRequest.SuggestConcepts = enabledSpaceEnablementField(false)
	trueResponse, err := handler.PutSpaceEnablements(t.Context(), &trueRequest, cm.PutSpaceEnablementsParams{
		SpaceID:            "space",
		XContentfulVersion: storedFalse.Sys.Version,
	})
	require.NoError(t, err)
	storedTrue := requireSpaceEnablementsSuccess(t, trueResponse)
	assert.Equal(t, storedFalse.Sys.Version+1, storedTrue.Sys.Version)
	assert.Equal(t, trueRequest.CrossSpaceLinks, storedTrue.CrossSpaceLinks)
	assert.Equal(t, trueRequest.SpaceTemplates, storedTrue.SpaceTemplates)
	assert.Equal(t, trueRequest.StudioExperiences, storedTrue.StudioExperiences)
	assert.Equal(t, trueRequest.SuggestConcepts, storedTrue.SuggestConcepts)
	assert.Equal(t, storedTrue, *handler.enablements["space"])
}

func newSpaceEnablementsTestHandler(t *testing.T) (*Handler, cm.SpaceEnablement) {
	t.Helper()

	handler := NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")

	initial := NewSpaceEnablementFromRequestFields("space", cm.SpaceEnablementData{
		CrossSpaceLinks:   enabledSpaceEnablementField(false),
		SpaceTemplates:    enabledSpaceEnablementField(false),
		StudioExperiences: enabledSpaceEnablementField(true),
		SuggestConcepts:   enabledSpaceEnablementField(false),
	})
	handler.enablements["space"] = &initial

	return handler, initial
}

func validSpaceEnablementData(enabled bool) cm.SpaceEnablementData {
	return cm.SpaceEnablementData{
		CrossSpaceLinks: enabledSpaceEnablementField(enabled),
		SpaceTemplates:  enabledSpaceEnablementField(enabled),
	}
}

func enabledSpaceEnablementField(enabled bool) cm.OptSpaceEnablementField {
	return cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: enabled})
}

func requireSpaceEnablementsError(t *testing.T, response cm.PutSpaceEnablementsRes, status int, expectedID string) {
	t.Helper()

	statusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, status, statusCode.StatusCode)
	errorResponse, ok := statusCode.Response.GetError()
	require.True(t, ok)
	assert.Equal(t, cm.ErrorSysTypeError, errorResponse.Sys.Type)
	assert.Equal(t, expectedID, errorResponse.Sys.ID)
	assert.NotEmpty(t, errorResponse.Message.Or(""))
}

func requireSpaceEnablementsSuccess(t *testing.T, response cm.PutSpaceEnablementsRes) cm.SpaceEnablement {
	t.Helper()

	statusCode, ok := response.(*cm.SpaceEnablementStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusOK, statusCode.StatusCode)

	return statusCode.Response
}
