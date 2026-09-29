package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesCursorPagination(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "role")
	for _, terminal := range []string{`"pages":{},`, `"pages":{"prev":"/spaces/space/roles?pagePrev=back"},`, ""} {
		t.Run(terminal, func(t *testing.T) {
			t.Parallel()

			queries := []string{"limit=100", "limit=100&pageNext=a%2Bb%2Fc%3D", "limit=100&pageNext=last"}
			pages := []string{
				`{"sys":{"type":"Array"},"limit":1,"items":[` + body + `],"pages":{"next":"/spaces/space/roles?pageNext=a%2Bb%2Fc%3D&limit=1"}}`,
				// An empty intermediate page with a new cursor must still be followed.
				`{"sys":{"type":"Array"},"items":[],"pages":{"next":"https://untrusted.invalid/spaces/space/roles?pageNext=last"}}`,
				`{"sys":{"type":"Array"},` + terminal + `"items":[` + strings.Replace(body, "Second", "Last arrival", 1) + `]}`,
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
	body := discoveryFixture(t, "role")
	first := `{"sys":{"type":"Array"},"items":[` + body + `],"pages":{"next":"/spaces/space/roles?pageNext=first"}}`

	tests := []struct {
		name   string
		page   string
		status int
		want   string
	}{
		{"empty next", `"pages":{"next":""}`, 200, "decode"},
		{"null next", `"pages":{"next":null}`, 200, "decode"},
		{"null pages", `"pages":null`, 200, "decode"},
		{"wrong next type", `"pages":{"next":42}`, 200, "decode"},
		{"malformed URL", `"pages":{"next":"%"}`, 200, "pages.next"},
		{"wrong space", `"pages":{"next":"/spaces/other/roles?pageNext=x"}`, 200, "endpoint"},
		{"wrong endpoint", `"pages":{"next":"/spaces/space/entries?pageNext=x"}`, 200, "endpoint"},
		{"userinfo", `"pages":{"next":"https://user:pass@host/spaces/space/roles?pageNext=x"}`, 200, "endpoint"},
		{"fragment", `"pages":{"next":"/spaces/space/roles?pageNext=x#part"}`, 200, "endpoint"},
		{"scheme", `"pages":{"next":"file:///spaces/space/roles?pageNext=x"}`, 200, "endpoint"},
		{"missing host", `"pages":{"next":"https:/spaces/space/roles?pageNext=x"}`, 200, "endpoint"},
		{"network relative", `"pages":{"next":"//host/spaces/space/roles?pageNext=x"}`, 200, "endpoint"},
		{"wrong item space", `"pages":{}`, 200, "Space link"},
		{"no cursor", `"pages":{"next":"/spaces/space/roles?limit=1"}`, 200, "pageNext"},
		{"empty cursor", `"pages":{"next":"/spaces/space/roles?pageNext="}`, 200, "pageNext"},
		{"duplicate cursor", `"pages":{"next":"/spaces/space/roles?pageNext=x&pageNext=y"}`, 200, "pageNext"},
		{"bad query", `"pages":{"next":"/spaces/space/roles?pageNext=%zz"}`, 200, "query"},
		{"backwards", `"pages":{"next":"/spaces/space/roles?pageNext=x&pagePrev=y"}`, 200, "pageNext"},
		{"offset link", `"pages":{"next":"/spaces/space/roles?pageNext=x&skip=2"}`, 200, "pageNext"},
		{"repeat", `"pages":{"next":"/spaces/space/roles?limit=42&pageNext=first"}`, 200, "repeats"},
		{"mixed", `"total":10,"pages":{}`, 200, "mixes"},
		{"offset after cursor", `"total":10,"skip":1`, 200, "mixes"},
		{"later denied", "", 403, "AccessDenied"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			count := 0
			response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++
				require.LessOrEqual(t, count, 2)

				if count == 1 {
					return discoveryHTTPResponse(request, 200, first), nil
				}

				page := `{"sys":{"type":"Array"},"items":[` + body + `],` + test.page + `}`
				if test.name == "wrong item space" {
					page = strings.Replace(page, `"id":"space"`, `"id":"other"`, 1)
				}

				if test.status == 403 {
					page = `{"sys":{"type":"Error","id":"AccessDenied"},"message":"denied"}`
				}

				return discoveryHTTPResponse(request, test.status, page), nil
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

		return discoveryHTTPResponse(request, 200, `{"sys":{"type":"Array"},"items":[],"pages":{"next":"/spaces/space/roles?pageNext=`+next+`"}}`), nil
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

				return discoveryHTTPResponse(request, 200, `{"sys":{"type":"Array"},"items":[],"pages":{"next":"/spaces/space/roles?pageNext=next"}}`), nil
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

			return discoveryHTTPResponse(request, 200, `{"sys":{"type":"Array"},"skip":0,"items":[`+discoveryFixture(t, "role")+`]}`), nil
		}

		assert.Equal(t, "limit=100&skip=1", request.URL.RawQuery)

		return discoveryHTTPResponse(request, 200, `{"sys":{"type":"Array"},"skip":1,"items":[]}`), nil
	}))
	require.Empty(t, response.Diagnostics)
	assert.Equal(t, 2, count)
}
