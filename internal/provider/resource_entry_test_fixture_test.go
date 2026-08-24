package provider_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

type entryFixtureErrorSink struct {
	mu  sync.Mutex
	err error
}

func (s *entryFixtureErrorSink) record(err error) {
	if err == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.err == nil {
		s.err = err
	}
}

func (s *entryFixtureErrorSink) error() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.err
}

type entryTestFixture struct {
	server    *cmt.Server
	recorder  *entryMutationRecorder
	errorSink *entryFixtureErrorSink
}

func newEntryAcceptanceFixture(t *testing.T) *entryTestFixture {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	errorSink := new(entryFixtureErrorSink)
	recorder := newEntryMutationRecorder(server, errorSink)

	return &entryTestFixture{server: server, recorder: recorder, errorSink: errorSink}
}

func managedEntryConfig(value string) string {
	return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = { managed = jsonencode({ "en-US" = %q }) }
}
`, value)
}

func getEntryFromTestServer(ctx context.Context, server *cmt.Server) (*cm.Entry, error) {
	response, err := server.Handler().GetEntry(ctx, cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	if err != nil {
		return nil, fmt.Errorf("get entry from test server: %w", err)
	}

	entry, ok := response.(*cm.Entry)
	if !ok {
		return nil, fmt.Errorf("%w: %T", errUnexpectedEntryResponseType, response)
	}

	return entry, nil
}

func getTestEntry(t *testing.T, server *cmt.Server) *cm.Entry {
	t.Helper()

	return getTestEntryForIDs(t, server, "space", "environment", "entry")
}

func getTestEntryForIDs(t *testing.T, server *cmt.Server, spaceID, environmentID, entryID string) *cm.Entry {
	t.Helper()

	response, err := server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: spaceID, EnvironmentID: environmentID, EntryID: entryID,
	})
	require.NoError(t, err)

	entry, ok := response.(*cm.Entry)
	require.True(t, ok)

	return entry
}
