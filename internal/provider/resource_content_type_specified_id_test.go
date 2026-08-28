package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

const (
	contentTypeTestPutPath      = "/spaces/space/environments/environment/content_types/content-type"
	contentTypeTestActivatePath = contentTypeTestPutPath + "/published"
)

type contentTypeMutationRequest struct {
	method         string
	path           string
	version        string
	versionPresent bool
	body           []byte
}

type contentTypeMutationRecorder struct {
	delegate http.Handler

	mu       sync.Mutex
	requests []contentTypeMutationRequest
	err      error
}

func (h *contentTypeMutationRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && request.URL.Path == contentTypeTestPutPath {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			h.recordError(err)
		} else {
			request.Body = io.NopCloser(bytes.NewReader(body))
			h.record(contentTypeMutationRequest{
				method: request.Method, path: request.URL.Path,
				version:        request.Header.Get("X-Contentful-Version"),
				versionPresent: len(request.Header.Values("X-Contentful-Version")) > 0,
				body:           body,
			})
		}
	}

	if request.Method == http.MethodPut && request.URL.Path == contentTypeTestActivatePath {
		h.record(contentTypeMutationRequest{
			method: request.Method, path: request.URL.Path,
			version:        request.Header.Get("X-Contentful-Version"),
			versionPresent: len(request.Header.Values("X-Contentful-Version")) > 0,
		})
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *contentTypeMutationRecorder) record(request contentTypeMutationRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = append(h.requests, request)
}

func (h *contentTypeMutationRecorder) recordError(err error) {
	if err == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.err == nil {
		h.err = err
	}
}

func (h *contentTypeMutationRecorder) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = nil
}

func (h *contentTypeMutationRecorder) snapshot() []contentTypeMutationRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]contentTypeMutationRequest(nil), h.requests...)
}

func (h *contentTypeMutationRecorder) handlerError() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.err
}

type contentTypeCollisionObservationAdapter struct {
	delegate  http.Handler
	errorSink *contentTypeMutationRecorder
}

// ServeHTTP turns an unsafe successful collision into the rejection the
// provider should have received. Delegating first makes the old version-fenced
// Create observable through the independent fake store: it overwrites the
// sentinel before Terraform sees the conflict, while create-only leaves it
// unchanged.
func (h *contentTypeCollisionObservationAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || request.URL.Path != contentTypeTestPutPath {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		message := "Version mismatch"
		h.errorSink.recordError(cmt.WriteContentfulManagementErrorResponse(
			responseWriter, http.StatusConflict, string(cm.ErrorSysIDVersionMismatch), &message, nil,
		))

		return
	}

	replayContentTypeAdapterResponse(responseWriter, recorder, h.errorSink)
}

func replayContentTypeAdapterResponse(responseWriter http.ResponseWriter, recorder *httptest.ResponseRecorder, errors *contentTypeMutationRecorder) {
	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, err := responseWriter.Write(recorder.Body.Bytes())
	errors.recordError(err)
}

func TestAccContentTypeResourceSpecifiedIDCreateUsesCreateOnlyRequest(t *testing.T) {
	t.Parallel()

	server, recorder := newContentTypeSpecifiedIDFixture(t)

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: managedContentTypeConfig("Created"),
			Check: func(*terraform.State) error {
				require.NoError(t, recorder.handlerError())

				draft, activation := requireContentTypePutThenActivate(t, recorder.snapshot())
				require.Equal(t, http.MethodPut, draft.method)
				require.Equal(t, contentTypeTestPutPath, draft.path)
				require.False(t, draft.versionPresent, "specified-ID Create must omit X-Contentful-Version")
				require.Empty(t, draft.version)
				require.JSONEq(t, `{
				  "name":"Created",
				  "description":"Managed Content Type",
				  "displayField":"title",
				  "fields":[{
				    "id":"title","name":"Title","type":"Symbol",
				    "localized":false,"disabled":false,"omitted":false,
				    "required":true,"validations":[]
				  }]
				}`, string(draft.body))

				require.Equal(t, http.MethodPut, activation.method)
				require.Equal(t, contentTypeTestActivatePath, activation.path)
				require.True(t, activation.versionPresent)
				require.Equal(t, "1", activation.version, "Create must activate the exact version returned by the draft PUT")

				contentType := getSpecifiedIDContentType(t, server)
				require.Equal(t, 2, contentType.Sys.Version)
				require.Equal(t, 1, contentType.Sys.PublishedVersion.Or(0))
				require.Equal(t, "Created", contentType.Name)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    managedContentTypeConfig("Created"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				require.NoError(t, recorder.handlerError())
				require.Empty(t, recorder.snapshot(), "a stable refresh and plan must not mutate the Content Type")

				return nil
			},
		},
	}})
}

func TestAccContentTypeResourceSpecifiedIDCollisionDoesNotMutateActivateOrAdopt(t *testing.T) {
	t.Parallel()

	server, recorder := newContentTypeSpecifiedIDFixture(t)
	server.SetContentType("space", "environment", "content-type", cm.ContentTypeRequestData{
		Name:         "Sentinel",
		Description:  cm.NewOptNilString("Must survive"),
		DisplayField: "sentinel",
		Fields: []cm.ContentTypeRequestDataFieldsItem{{
			ID: "sentinel", Name: "Sentinel", Type: "Symbol",
			Localized: cm.NewOptBool(false), Required: cm.NewOptBool(true),
			Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false),
		}},
	})
	beforeCollision := snapshotSpecifiedIDContentType(t, server)

	adapter := &contentTypeCollisionObservationAdapter{delegate: recorder, errorSink: recorder}

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:      managedContentTypeConfig("Must not replace sentinel"),
			ExpectError: regexp.MustCompile(`(?s)Failed to create content type.*VersionMismatch.*version precondition was not\s+satisfied`),
		},
		{
			PreConfig: func() {
				require.NoError(t, recorder.handlerError())

				contentType := getSpecifiedIDContentType(t, server)
				require.JSONEq(t, string(beforeCollision), string(snapshotSpecifiedIDContentType(t, server)), "collision must preserve the complete remote Content Type")
				require.Equal(t, 1, contentType.Sys.Version, "collision must not advance the sentinel version")
				require.False(t, contentType.Sys.PublishedVersion.IsSet(), "collision must not activate the sentinel")
				require.Equal(t, "Sentinel", contentType.Name)
				require.Equal(t, "Must survive", contentType.Description.Or(""))
				require.Equal(t, "sentinel", contentType.DisplayField.Or(""))
				require.Len(t, contentType.Fields, 1)
				require.Equal(t, "sentinel", contentType.Fields[0].ID)

				requests := recorder.snapshot()
				require.Len(t, requests, 1, "collision must not be followed by activation")
				require.Equal(t, contentTypeTestPutPath, requests[0].path)
				require.False(t, requests[0].versionPresent, "collision request must be create-only")
				require.Empty(t, requests[0].version, "collision request must be create-only")
			},
			Config:      managedContentTypeConfig("Must not replace sentinel"),
			ExpectError: regexp.MustCompile(`(?s)Failed to create content type.*VersionMismatch.*version precondition was not\s+satisfied`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
			}},
		},
	}})
}

func newContentTypeSpecifiedIDFixture(t *testing.T) (*cmt.Server, *contentTypeMutationRecorder) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	return server, &contentTypeMutationRecorder{delegate: server}
}

func managedContentTypeConfig(name string) string {
	return fmt.Sprintf(`
resource "contentful_content_type" "test" {
  space_id        = "space"
  environment_id  = "environment"
  content_type_id = "content-type"

  name          = %q
  description   = "Managed Content Type"
  display_field = "title"

  fields = [{
    id        = "title"
    name      = "Title"
    type      = "Symbol"
    localized = false
    required  = true
  }]
}
`, name)
}

func getSpecifiedIDContentType(t *testing.T, server *cmt.Server) *cm.ContentType {
	t.Helper()

	response, err := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
	})
	require.NoError(t, err)

	contentType, ok := response.(*cm.ContentType)
	require.True(t, ok)

	return contentType
}

func snapshotSpecifiedIDContentType(t *testing.T, server *cmt.Server) []byte {
	t.Helper()

	snapshot, err := json.Marshal(getSpecifiedIDContentType(t, server))
	require.NoError(t, err)

	return snapshot
}

func requireContentTypePutThenActivate(t *testing.T, requests []contentTypeMutationRequest) (contentTypeMutationRequest, contentTypeMutationRequest) {
	t.Helper()

	require.Len(t, requests, 2, "expected one Content Type draft PUT followed by one activation")
	require.Equal(t, contentTypeTestPutPath, requests[0].path)
	require.Equal(t, contentTypeTestActivatePath, requests[1].path)

	return requests[0], requests[1]
}
