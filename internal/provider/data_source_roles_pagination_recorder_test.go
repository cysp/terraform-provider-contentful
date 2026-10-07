package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// Observe successful pages only: a retried request is not another page.
type rolePaginationRecorder struct {
	transport http.RoundTripper
	path      string
	mu        sync.Mutex
	limit     int
	observed  []rolePaginationPage
}

type rolePaginationPage struct {
	query url.Values
	ids   []string
	total *int
	skip  *int
	next  string
}

func (recorder *rolePaginationRecorder) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet || request.URL.Path != recorder.path {
		return recorder.transport.RoundTrip(request) //nolint:wrapcheck
	}

	recorder.mu.Lock()
	limit := recorder.limit
	recorder.mu.Unlock()

	request = request.Clone(request.Context())
	query := request.URL.Query()

	if limit != 0 {
		query.Set("limit", strconv.Itoa(limit))
		request.URL.RawQuery = query.Encode()
	}

	response, err := recorder.transport.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("read live Role page: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return response, nil
	}

	body, err := io.ReadAll(response.Body)
	response.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("read live Role page body: %w", err)
	}

	// Restore the exact response bytes for the real generated client to decode.
	response.Body = io.NopCloser(bytes.NewReader(body))

	var collection struct {
		Total *int `json:"total"`
		Skip  *int `json:"skip"`
		Items []struct {
			Sys struct {
				ID string `json:"id"`
			} `json:"sys"`
		} `json:"items"`
		Pages struct {
			Next string `json:"next"`
		} `json:"pages"`
	}

	err = json.Unmarshal(body, &collection)
	if err != nil {
		return nil, fmt.Errorf("decode live Role page: %w", err)
	}

	page := rolePaginationPage{query: query, total: collection.Total, skip: collection.Skip, next: collection.Pages.Next}
	for _, item := range collection.Items {
		page.ids = append(page.ids, item.Sys.ID)
	}

	recorder.mu.Lock()
	recorder.observed = append(recorder.observed, page)
	recorder.mu.Unlock()

	return response, nil
}

func (recorder *rolePaginationRecorder) setLimit(limit int) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.limit = limit
	recorder.observed = nil
}

func (recorder *rolePaginationRecorder) pages() []rolePaginationPage {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return append([]rolePaginationPage(nil), recorder.observed...)
}

func requireRolePaginationBaseline(t *testing.T, pages []rolePaginationPage) []string {
	t.Helper()
	require.NotEmpty(t, pages, "baseline must read the live Role collection")

	baseline := pages[0].ids
	require.GreaterOrEqual(t, len(baseline), 2, "live pagination requires at least two existing Roles")

	seen := make(map[string]bool, len(baseline))
	for _, id := range baseline {
		require.NotEmpty(t, id)
		require.NotContains(t, seen, id, "baseline must contain distinct Roles")
		seen[id] = true
	}

	for _, page := range pages {
		require.Equal(t, url.Values{"limit": {"100"}}, page.query, "baseline must fit in one default-sized page per read")
		require.Empty(t, page.next, "baseline must have no cursor continuation")
		require.Equal(t, baseline, page.ids, "baseline changed between reads")

		if page.total != nil {
			require.Equal(t, len(baseline), *page.total, "baseline must contain the complete collection")
		}
	}

	return baseline
}

func requireRolePaginationReads(t *testing.T, pages []rolePaginationPage, baseline []string) {
	t.Helper()
	require.NotEmpty(t, pages, "limited read must request the live Role collection")

	initial := url.Values{"limit": {"1"}}
	nextQuery := initial

	var (
		ids                           []string
		nonemptyPages, completedReads int
	)

	for _, page := range pages {
		require.Equal(t, nextQuery, page.query, "each read must follow its previous page's continuation")
		require.LessOrEqual(t, len(page.ids), 1, "Contentful must honor the one-Role page limit")

		ids = append(ids, page.ids...)
		if len(page.ids) != 0 {
			nonemptyPages++
		}

		nextQuery = initial
		protocol := "cursor"

		if page.total != nil {
			protocol = "offset"

			require.NotNil(t, page.skip)
			require.Equal(t, len(ids)-len(page.ids), *page.skip)
			require.Empty(t, page.next)
			require.Equal(t, len(baseline), *page.total)

			if len(ids) < *page.total {
				nextQuery = url.Values{"limit": {"1"}, "skip": {strconv.Itoa(len(ids))}}
			}
		} else if page.next != "" {
			navigation, err := url.Parse(page.next)
			require.NoError(t, err)

			cursor := navigation.Query().Get("pageNext")
			require.NotEmpty(t, cursor)
			nextQuery = url.Values{"limit": {"1"}, "pageNext": {cursor}}
		}

		if nextQuery.Has("skip") || nextQuery.Has("pageNext") {
			continue
		}

		require.GreaterOrEqual(t, nonemptyPages, 2, "each completed read must contain multiple successful nonempty pages")
		require.Equal(t, baseline, ids, "each completed traversal must match the independent single-page baseline")
		t.Logf("live Role pagination: %s, %d nonempty pages, %d Roles", protocol, nonemptyPages, len(ids))

		completedReads++
		ids = nil
		nonemptyPages = 0
	}

	require.Equal(t, initial, nextQuery, "last observed traversal must reach a terminal response")
	require.Positive(t, completedReads)
}
