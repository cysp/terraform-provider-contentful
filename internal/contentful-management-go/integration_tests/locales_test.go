package integration_tests_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleCRUD(t *testing.T) {
	t.Parallel()

	_, testserver := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	createResponse, err := client.PutLocale(t.Context(), &cm.LocaleRequest{
		Name:         "English (US)",
		Code:         "en-US",
		FallbackCode: cm.NewOptNilStringNull(),
	}, cm.PutLocaleParams{
		SpaceID:  "0p38pssr0fi3",
		LocaleID: "en-US",
	})
	require.NoError(t, err)

	var localeVersion int

	switch createResponse := createResponse.(type) {
	case *cm.LocaleStatusCode:
		assert.Equal(t, http.StatusCreated, createResponse.GetStatusCode())
		assert.Equal(t, "en-US", createResponse.Response.Sys.ID)
		assert.Equal(t, "English (US)", createResponse.Response.Name)
		localeVersion = createResponse.Response.Sys.Version
	default:
		t.Fatalf("unexpected create response type %T", createResponse)
	}

	readResponse, err := client.GetLocale(t.Context(), cm.GetLocaleParams{
		SpaceID:  "0p38pssr0fi3",
		LocaleID: "en-US",
	})
	require.NoError(t, err)

	switch readResponse := readResponse.(type) {
	case *cm.Locale:
		assert.Equal(t, "en-US", readResponse.Sys.ID)
		assert.Equal(t, "English (US)", readResponse.Name)
	default:
		t.Fatalf("unexpected get response type %T", readResponse)
	}

	updateParams := cm.PutLocaleParams{
		SpaceID:  "0p38pssr0fi3",
		LocaleID: "en-US",
	}
	updateParams.XContentfulVersion.SetTo(localeVersion)

	updateResponse, err := client.PutLocale(t.Context(), &cm.LocaleRequest{
		Name:         "English (US) Updated",
		Code:         "en-US",
		Optional:     cm.NewOptBool(true),
		FallbackCode: cm.NewOptNilString("en-GB"),
	}, updateParams)
	require.NoError(t, err)

	switch updateResponse := updateResponse.(type) {
	case *cm.LocaleStatusCode:
		assert.Equal(t, http.StatusOK, updateResponse.GetStatusCode())
		assert.Equal(t, "English (US) Updated", updateResponse.Response.Name)
		fallbackCode, fallbackCodeSet := updateResponse.Response.FallbackCode.Get()
		require.True(t, fallbackCodeSet)
		assert.Equal(t, "en-GB", fallbackCode)
		assert.True(t, updateResponse.Response.Optional)
		localeVersion = updateResponse.Response.Sys.Version
	default:
		t.Fatalf("unexpected update response type %T", updateResponse)
	}

	deleteParams := cm.DeleteLocaleParams{
		SpaceID:  "0p38pssr0fi3",
		LocaleID: "en-US",
	}
	deleteParams.XContentfulVersion.SetTo(localeVersion)

	deleteResponse, err := client.DeleteLocale(t.Context(), deleteParams)
	require.NoError(t, err)

	switch deleteResponse := deleteResponse.(type) {
	case *cm.NoContent:
		require.NotNil(t, deleteResponse)
	default:
		t.Fatalf("unexpected delete response type %T", deleteResponse)
	}
}

func TestGetLocaleNotFound(t *testing.T) {
	t.Parallel()

	_, testserver := testContentfulManagementHTTPTestServer(t)
	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	response, err := client.GetLocale(t.Context(), cm.GetLocaleParams{
		SpaceID:  "0p38pssr0fi3",
		LocaleID: "nonexistent",
	})
	require.NoError(t, err)

	statusCodeResponse, ok := response.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, statusCodeResponse.GetStatusCode())
}
