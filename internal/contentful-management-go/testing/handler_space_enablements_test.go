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

	for name, enabled := range map[string]bool{
		"disabled": false,
		"enabled":  true,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler, before := newSpaceEnablementsTestHandler(t)
			request := validSpaceEnablementData(enabled)

			response, err := handler.PutSpaceEnablements(t.Context(), &request, cm.PutSpaceEnablementsParams{
				SpaceID:            "space",
				XContentfulVersion: before.Sys.Version,
			})
			require.NoError(t, err)
			stored := requireSpaceEnablementsSuccess(t, response)
			assert.Equal(t, before.Sys.Version+1, stored.Sys.Version)
			assert.Equal(t, request.CrossSpaceLinks, stored.CrossSpaceLinks)
			assert.Equal(t, request.SpaceTemplates, stored.SpaceTemplates)
			assert.Equal(t, stored, *handler.enablements["space"])
		})
	}
}

func newSpaceEnablementsTestHandler(t *testing.T) (*Handler, cm.SpaceEnablement) {
	t.Helper()

	handler := NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")

	initial := NewSpaceEnablementFromRequestFields("space", validSpaceEnablementData(false))
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
