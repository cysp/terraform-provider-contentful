package provider_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

// entryCollisionObservationAdapter converts an unsafe successful collision
// response into the rejection the provider should have received. Delegating
// first makes an incorrect version-fenced Create observable through the
// independent remote store: such an implementation mutates the sentinel before
// receiving the rejection, while a create-only request is rejected unchanged.
type entryCollisionObservationAdapter struct {
	delegate  http.Handler
	errorSink *entryFixtureErrorSink
}

func (h *entryCollisionObservationAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || request.URL.Path != entryTestUpdatePath {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		message := "Version mismatch"
		h.errorSink.record(cmt.WriteContentfulManagementErrorResponse(
			responseWriter, http.StatusConflict, string(cm.ErrorSysIDVersionMismatch), &message, nil,
		))

		return
	}

	replayEntryAdapterResponse(responseWriter, recorder, h.errorSink)
}

func TestAccEntryResourceSpecifiedIDCreateUsesCreateOnlyRequest(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: managedEntryConfig("created"),
				Check: func(*terraform.State) error {
					require.NoError(t, fixture.recorder.handlerError())

					draft, publish := requireEntryUpdateThenPublish(t, fixture.recorder.snapshot())
					require.Equal(t, http.MethodPut, draft.method)
					require.Equal(t, entryTestUpdatePath, draft.path)
					require.Empty(t, draft.versionValues)
					require.Equal(t, []string{"article"}, draft.contentTypeValues)
					require.JSONEq(t, `{"fields":{"managed":{"en-US":"created"}},"metadata":{"concepts":[],"tags":[]}}`, string(draft.body))

					requireEntryPublish(t, publish, entryTestPublishPath)
					require.Equal(t, "1", publish.version, "Create must publish the version returned by the draft PUT")

					entry := getTestEntry(t, fixture.server)
					require.Equal(t, 2, entry.Sys.Version)
					require.Equal(t, 1, entry.Sys.PublishedVersion.Or(0))
					require.JSONEq(t, `{"en-US":"created"}`, string(entry.Fields.Value["managed"]))

					return nil
				},
			},
			{
				PreConfig: fixture.recorder.reset,
				Config:    managedEntryConfig("created"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					require.NoError(t, fixture.recorder.handlerError())
					requireNoEntryMutations(t, fixture.recorder)

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceSpecifiedIDCollisionDoesNotMutateOrAdopt(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	fixture.server.SetEntry("space", "environment", "sentinel-type", "entry", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"sentinel": jx.Raw(`{"en-US":"must survive"}`),
		}),
	})

	adapter := &entryCollisionObservationAdapter{delegate: fixture.recorder, errorSink: fixture.errorSink}

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      managedEntryConfig("must not replace sentinel"),
				ExpectError: regexp.MustCompile(`(?s)Failed to create entry.*VersionMismatch.*version precondition was not\s+satisfied`),
			},
			{
				PreConfig: func() {
					require.NoError(t, fixture.recorder.handlerError())

					entry := getTestEntry(t, fixture.server)
					require.Equal(t, 1, entry.Sys.Version, "collision must not advance the sentinel version")
					require.False(t, entry.Sys.PublishedVersion.IsSet(), "collision must not publish the sentinel")
					require.Equal(t, "sentinel-type", entry.Sys.ContentType.Sys.ID)
					require.Len(t, entry.Fields.Value, 1)
					require.JSONEq(t, `{"en-US":"must survive"}`, string(entry.Fields.Value["sentinel"]))
					require.NotContains(t, entry.Fields.Value, "managed")

					requests := fixture.recorder.snapshot()
					require.Len(t, requests, 1, "collision must not be followed by Publish")
					requireEntryUpdate(t, requests[0])
					require.Empty(t, requests[0].version)
					require.Empty(t, requests[0].versionValues, "collision request must be create-only")
				},
				Config:      managedEntryConfig("must not replace sentinel"),
				ExpectError: regexp.MustCompile(`(?s)Failed to create entry.*VersionMismatch.*version precondition was not\s+satisfied`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionCreate),
				}},
			},
		},
	})
}
