package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourcePublishedDestroyUsesUnpublishResponseVersion(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{PreConfig: fixture.recorder.reset, Config: managedEntryConfig("one"), Destroy: true},
		},
	})

	requests := fixture.recorder.destructiveSnapshot()
	require.Len(t, requests, 2)
	requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
	requireEntryDestructiveRequest(t, requests[1], entryTestUpdatePath, "3")
	require.Empty(t, fixture.recorder.readSnapshot(), "stored private version must not be replaced by a destroy-time GET")
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourcePublishedDestroyUsesExactHigherUnpublishResponseVersion(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	fault := &entryPostResponseUpdateAdapter{
		delegate: fixture.server, server: fixture.server,
		method: http.MethodDelete, path: entryTestPublishPath, updates: 2,
		returnUpdatedEntry: true, errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = fault
	fault.shot.arm()

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{PreConfig: fixture.recorder.reset, Config: managedEntryConfig("one"), Destroy: true},
		},
	})

	requests := fixture.recorder.destructiveSnapshot()
	require.Len(t, requests, 2)
	requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
	requireEntryDestructiveRequest(t, requests[1], entryTestUpdatePath, "5")
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourcePublishedDestroyRequiresEntryUnpublishResponse(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	fault := &entryNoContentUnpublishAdapter{delegate: fixture.server}
	fixture.recorder.delegate = fault
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					fault.shot.arm()
					fixture.recorder.reset()
				},
				Config:      managedEntryConfig("one"),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Failed to unpublish entry`),
			},
			{
				PreConfig: func() {
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 1, "delete must not follow an unpublish response without an Entry version")
					requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
					entry := getTestEntry(t, fixture.server)
					require.Equal(t, 2, entry.Sys.Version)
					require.True(t, entry.Sys.PublishedVersion.IsSet())
					fixture.recorder.reset()
				},
				Config:  managedEntryConfig("one"),
				Destroy: true,
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourcePublishedStaleDestroyStopsAfterUnpublish(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					entry, err := advanceTestEntry(t.Context(), fixture.server)
					require.NoError(t, err)
					require.Equal(t, 3, entry.Sys.Version)
					fixture.recorder.reset()
				},
				Config:      managedEntryConfig("one"),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`version precondition was not\s+satisfied`),
			},
			{
				PreConfig: func() {
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 1, "delete must not follow a stale unpublish")
					requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
					require.Equal(t, 3, getTestEntry(t, fixture.server).Sys.Version)
					require.Empty(t, fixture.recorder.readSnapshot(), "stale private authority must not be refreshed during destroy")

					options.Plan.NoRefresh = false

					fixture.recorder.reset()
				},
				Config:  managedEntryConfig("one"),
				Destroy: true,
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourceUnpublishedStaleDestroyUsesStoredVersion(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	options := &resource.AdditionalCLIOptions{}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					entry, err := unpublishTestEntry(t.Context(), fixture.server)
					require.NoError(t, err)
					require.Equal(t, 3, entry.Sys.Version)
					fixture.recorder.reset()
				},
				Config: managedEntryConfig("one"),
			},
			{
				PreConfig: func() {
					entry, err := advanceTestEntry(t.Context(), fixture.server)
					require.NoError(t, err)
					require.Equal(t, 4, entry.Sys.Version)

					options.Plan.NoRefresh = true

					fixture.recorder.reset()
				},
				Config:      managedEntryConfig("one"),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`version precondition was not\s+satisfied`),
			},
			{
				PreConfig: func() {
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 1)
					requireEntryDestructiveRequest(t, requests[0], entryTestUpdatePath, "3")
					require.Equal(t, 4, getTestEntry(t, fixture.server).Sys.Version)
					require.Empty(t, fixture.recorder.readSnapshot(), "stale private authority must not be refreshed during destroy")

					options.Plan.NoRefresh = false

					fixture.recorder.reset()
				},
				Config:  managedEntryConfig("one"),
				Destroy: true,
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourceConcurrentChangeBetweenUnpublishAndDelete(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	fault := &entryPostResponseUpdateAdapter{
		delegate: fixture.server, server: fixture.server,
		method: http.MethodDelete, path: entryTestPublishPath, updates: 1,
		errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = fault
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					fault.shot.arm()
					fixture.recorder.reset()
				},
				Config:      managedEntryConfig("one"),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`version precondition was not\s+satisfied`),
			},
			{
				PreConfig: func() {
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 2)
					requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
					requireEntryDestructiveRequest(t, requests[1], entryTestUpdatePath, "3")
					require.Equal(t, 4, getTestEntry(t, fixture.server).Sys.Version)

					options.Plan.NoRefresh = false

					fixture.recorder.reset()
				},
				Config:  managedEntryConfig("one"),
				Destroy: true,
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourceTaintedReplacementFetchesMissingDeleteVersion(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					entry, err := advanceTestEntry(t.Context(), fixture.server)
					require.NoError(t, err)
					require.Equal(t, 3, entry.Sys.Version)
					fixture.recorder.reset()
				},
				Config: managedEntryConfig("one"),
				Taint:  []string{"contentful_entry.test"},
				Check: func(*terraform.State) error {
					require.Len(t, fixture.recorder.readSnapshot(), 1, "only missing private state may authorize a destroy-time GET")
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 2)
					requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "3")
					requireEntryDestructiveRequest(t, requests[1], entryTestUpdatePath, "4")

					return nil
				},
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourceTaintedReplacementPreservesFetchedDeleteVersion(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	fault := &entryPostResponseUpdateAdapter{
		delegate: fixture.server, server: fixture.server,
		method: http.MethodGet, path: entryTestUpdatePath, updates: 1,
		errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = fault
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					fault.shot.arm()
					fixture.recorder.reset()
				},
				Config:      managedEntryConfig("one"),
				Taint:       []string{"contentful_entry.test"},
				ExpectError: regexp.MustCompile(`version precondition was not\s+satisfied`),
			},
			{
				PreConfig: func() {
					require.Len(t, fixture.recorder.readSnapshot(), 1)
					requests := fixture.recorder.destructiveSnapshot()
					require.Len(t, requests, 1, "delete must not follow a stale fetched-version unpublish")
					requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
					require.Equal(t, 3, getTestEntry(t, fixture.server).Sys.Version)
					fixture.recorder.reset()
				},
				Config: managedEntryConfig("one"),
			},
		},
	})
}

//nolint:paralleltest // Mocked acceptance tests parallelize only when TF_ACC_MOCKED is set.
func TestAccEntryResourceDestroyTreatsOnlyNotFoundAsAbsent(t *testing.T) {
	parallelWhenMocked(t)

	fixture := newEntryAcceptanceFixture(t)
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					response, err := fixture.server.Handler().DeleteEntry(t.Context(), cm.DeleteEntryParams{
						SpaceID:            "space",
						EnvironmentID:      "environment",
						EntryID:            "entry",
						XContentfulVersion: cm.NewOptInt(2),
					})
					require.NoError(t, err)
					require.IsType(t, &cm.NoContent{}, response)
					fixture.recorder.reset()
				},
				Config:  managedEntryConfig("one"),
				Destroy: true,
			},
		},
	})

	requests := fixture.recorder.destructiveSnapshot()
	require.Len(t, requests, 2)
	requireEntryDestructiveRequest(t, requests[0], entryTestPublishPath, "2")
	requireEntryDestructiveRequest(t, requests[1], entryTestUpdatePath, "2")
	require.Equal(t, http.StatusNotFound, entryResponseStatus(t, fixture.server))
}

func entryResponseStatus(t *testing.T, server *cmt.Server) int {
	t.Helper()

	response, err := server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	status, ok := response.(cm.StatusCodeResponse)
	require.True(t, ok)

	return status.GetStatusCode()
}
