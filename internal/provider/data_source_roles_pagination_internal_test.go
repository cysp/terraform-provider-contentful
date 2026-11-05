package provider

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesCursorPagination(t *testing.T) {
	t.Parallel()

	body := testJSONObject(discoveryFixture(t, "role"))
	last := maps.Clone(body)
	last["name"] = "Last arrival"

	for _, terminal := range []struct {
		name    string
		members map[string]any
	}{
		{"empty pages", map[string]any{"pages": map[string]any{}}},
		{"empty prev", map[string]any{"pages": map[string]any{"prev": ""}}},
		{"prev", map[string]any{"pages": map[string]any{"prev": "/spaces/space/roles?pagePrev=back"}}},
		{"absent pages", map[string]any{}},
	} {
		t.Run(terminal.name, func(t *testing.T) {
			t.Parallel()

			final := map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{last}}
			maps.Copy(final, terminal.members)

			queries := []string{"limit=100", "limit=100&pageNext=a%2Bb%2Fc%3D", "limit=100&pageNext=last"}
			pages := []string{
				testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "limit": 1, "items": []any{body}, "pages": map[string]any{"next": "/spaces/space/roles?pageNext=a%2Bb%2Fc%3D&limit=1", "prev": ""}}),
				// An empty intermediate page with a new cursor must still be followed.
				testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}, "pages": map[string]any{"next": "https://untrusted.invalid/spaces/space/roles?pageNext=last"}}),
				testJSON(final),
			}
			count := 0
			response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				require.Less(t, count, len(pages))
				assert.Equal(t, "https", request.URL.Scheme)
				assert.Equal(t, "example.invalid", request.URL.Host)
				assert.Equal(t, "/base/spaces/space/roles", request.URL.Path)
				assert.Equal(t, "Bearer synthetic-token", request.Header.Get("Authorization"))
				assert.Equal(t, queries[count], request.URL.RawQuery)
				page := pages[count]
				count++

				return discoveryHTTPResponse(request, 200, page), nil
			}))
			require.Empty(t, response.Diagnostics)
			assert.Equal(t, 3, count)

			var data RolesDataSourceModel
			require.Empty(t, response.State.Get(t.Context(), &data))
			require.Len(t, data.Roles, 2)
			assert.Equal(t, "Second", data.Roles[0].Name.ValueString())
			assert.Equal(t, "Last arrival", data.Roles[1].Name.ValueString())
			assert.Equal(t, data.Roles[0].RoleID, data.Roles[1].RoleID)
		})
	}
}

func TestRolesCursorPaginationErrorsDoNotPublish(t *testing.T) {
	t.Parallel()
	body := testJSONObject(discoveryFixture(t, "role"))
	first := testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{body}, "pages": map[string]any{"next": "/spaces/space/roles?pageNext=first"}})
	wrongSpace := testJSONObject(discoveryFixture(t, "role"))
	wrongSpace["sys"] = map[string]any{
		"id": "item-b", "type": "Role", "version": 7,
		"space": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "other"}},
	}

	// Synthetic credentials exercise rejection of response-provided userinfo.
	const userinfoURL = "https://user:pass@host/spaces/space/roles?pageNext=x" //nolint:gosec

	tests := []struct {
		name   string
		page   map[string]any
		status int
		want   string
		first  string
		query  string
	}{
		{name: "empty next", page: map[string]any{"pages": map[string]any{"next": ""}}, status: 200, want: "pages.next"},
		{name: "null next", page: map[string]any{"pages": map[string]any{"next": nil}}, status: 200, want: "decode"},
		{name: "null pages", page: map[string]any{"pages": nil}, status: 200, want: "decode"},
		{name: "wrong next type", page: map[string]any{"pages": map[string]any{"next": 42}}, status: 200, want: "decode"},
		{name: "malformed URL", page: map[string]any{"pages": map[string]any{"next": "%"}}, status: 200, want: "pages.next"},
		{name: "wrong space", page: map[string]any{"pages": map[string]any{"next": "/spaces/other/roles?pageNext=x"}}, status: 200, want: "endpoint"},
		{name: "wrong endpoint", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/entries?pageNext=x"}}, status: 200, want: "endpoint"},
		{name: "userinfo", page: map[string]any{"pages": map[string]any{"next": userinfoURL}}, status: 200, want: "endpoint"},
		{name: "fragment", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=x#part"}}, status: 200, want: "endpoint"},
		{name: "scheme", page: map[string]any{"pages": map[string]any{"next": "file:///spaces/space/roles?pageNext=x"}}, status: 200, want: "endpoint"},
		{name: "missing host", page: map[string]any{"pages": map[string]any{"next": "https:/spaces/space/roles?pageNext=x"}}, status: 200, want: "endpoint"},
		{name: "network relative", page: map[string]any{"pages": map[string]any{"next": "//host/spaces/space/roles?pageNext=x"}}, status: 200, want: "endpoint"},
		{name: "wrong item space", page: map[string]any{"pages": map[string]any{}, "items": []any{wrongSpace}}, status: 200, want: "Space link"},
		{name: "no cursor", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?limit=1"}}, status: 200, want: "pageNext"},
		{name: "empty cursor", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext="}}, status: 200, want: "pageNext"},
		{name: "duplicate cursor", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=x&pageNext=y"}}, status: 200, want: "pageNext"},
		{name: "bad query", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=%zz"}}, status: 200, want: "query"},
		{name: "backwards", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=x&pagePrev=y"}}, status: 200, want: "pageNext"},
		{name: "offset link", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=x&skip=2"}}, status: 200, want: "pageNext"},
		{name: "repeat", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?limit=42&pageNext=first"}}, status: 200, want: "repeats"},
		{name: "mixed", page: map[string]any{"total": 10, "pages": map[string]any{}}, status: 200, want: "mixes"},
		{name: "cursor after offset", page: map[string]any{"pages": map[string]any{"next": "/spaces/space/roles?pageNext=next"}}, status: 200, want: "mixes", first: testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "skip": 0, "total": 2, "items": []any{body}}), query: "limit=100&skip=1"},
		{name: "offset after cursor", page: map[string]any{"total": 10, "skip": 1}, status: 200, want: "mixes"},
		{name: "later denied", page: map[string]any{"sys": map[string]any{"type": "Error", "id": "AccessDenied"}, "message": "denied"}, status: 403, want: "AccessDenied"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			count := 0
			response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++
				require.LessOrEqual(t, count, 2)

				if count == 1 {
					assert.Equal(t, "limit=100", request.URL.RawQuery)

					if test.first != "" {
						return discoveryHTTPResponse(request, 200, test.first), nil
					}

					return discoveryHTTPResponse(request, 200, first), nil
				}

				expectedQuery := "limit=100&pageNext=first"
				if test.query != "" {
					expectedQuery = test.query
				}

				assert.Equal(t, expectedQuery, request.URL.RawQuery)

				page := map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{body}}
				maps.Copy(page, test.page)

				if test.status != http.StatusOK {
					page = test.page
				}

				return discoveryHTTPResponse(request, test.status, testJSON(page)), nil
			}))
			require.True(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Contains(t, fmt.Sprint(response.Diagnostics), test.want)
			assert.True(t, response.State.Raw.IsNull())
			assert.Equal(t, 2, count)
		})
	}
}

func TestRolesCursorCycle(t *testing.T) {
	t.Parallel()

	count := 0
	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Less(t, count, 3)
		next := []string{"one", "two", "one"}[count]
		count++

		return discoveryHTTPResponse(request, 200, testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}, "pages": map[string]any{"next": "/spaces/space/roles?pageNext=" + next}})), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.Contains(t, fmt.Sprint(response.Diagnostics), "repeats")
	assert.True(t, response.State.Raw.IsNull())
	assert.Equal(t, 3, count)
}

func TestRolesCursorOperationDeadline(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		started := time.Now()
		count := 0
		inputs := map[string]any{"space_id": "space", "timeouts": map[string]tftypes.Value{"read": tftypes.NewValue(tftypes.String, "1s")}}
		response := discoveryReadTest(t.Context(), t, NewRolesDataSource, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			count++
			deadline, ok := request.Context().Deadline()
			require.True(t, ok)
			assert.Equal(t, started.Add(time.Second), deadline)

			if count == 1 {
				time.Sleep(750 * time.Millisecond)

				return discoveryHTTPResponse(request, 200, testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}, "pages": map[string]any{"next": "/spaces/space/roles?pageNext=next"}})), nil
			}

			<-request.Context().Done()

			return nil, context.DeadlineExceeded
		}))
		require.True(t, response.Diagnostics.HasError())
		assert.True(t, response.State.Raw.IsNull())
		assert.Equal(t, 2, count)
		assert.Equal(t, time.Second, time.Since(started))
	})
}

func TestRolesOffsetWithoutTotal(t *testing.T) {
	t.Parallel()

	count := 0
	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		count++
		require.LessOrEqual(t, count, 2)

		if count == 1 {
			assert.Equal(t, "limit=100", request.URL.RawQuery)

			return discoveryHTTPResponse(request, 200, testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "skip": 0, "items": []any{testJSONObject(discoveryFixture(t, "role"))}})), nil
		}

		assert.Equal(t, "limit=100&skip=1", request.URL.RawQuery)

		return discoveryHTTPResponse(request, 200, testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "skip": 1, "items": []any{}})), nil
	}))
	require.Empty(t, response.Diagnostics)
	assert.Equal(t, 2, count)
}

func TestRolesOffsetPaginationRetainsModeWithoutMetadata(t *testing.T) {
	t.Parallel()

	body := testJSONObject(discoveryFixture(t, "role"))
	middle := maps.Clone(body)
	middle["name"] = "Middle"
	last := maps.Clone(body)
	last["name"] = "Last"
	pages := []string{
		testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "total": 3, "skip": 0, "items": []any{body}}),
		testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{middle}}),
		testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{last}}),
		testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}}),
	}
	queries := []string{"limit=100", "limit=100&skip=1", "limit=100&skip=2", "limit=100&skip=3"}
	count := 0
	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Less(t, count, len(pages))
		assert.Equal(t, queries[count], request.URL.RawQuery)
		page := pages[count]
		count++

		return discoveryHTTPResponse(request, 200, page), nil
	}))
	require.Empty(t, response.Diagnostics)
	assert.Equal(t, 4, count)

	var data RolesDataSourceModel
	require.Empty(t, response.State.Get(t.Context(), &data))
	require.Len(t, data.Roles, 3)
	assert.Equal(t, "Second", data.Roles[0].Name.ValueString())
	assert.Equal(t, "Middle", data.Roles[1].Name.ValueString())
	assert.Equal(t, "Last", data.Roles[2].Name.ValueString())
}
