package cmtesting_test

import (
	"encoding/base64"
	"math"
	"net/http/httptest"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRolesPagination(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	for _, id := range []string{"c", "a", "b"} {
		server.SetRole("space", id, cm.RoleData{Name: id, Permissions: cm.RoleDataPermissions{}, Policies: []cm.RoleDataPoliciesItem{}})
	}
	// Use the generated HTTP boundary to cover both query directions and decoding.
	httpServer := httptest.NewServer(server)
	t.Cleanup(httpServer.Close)
	client, err := cm.NewClient(httpServer.URL, cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken), cm.WithClient(cm.NewTransportClient(httpServer.Client(), "")))
	require.NoError(t, err)

	tests := []struct {
		name   string
		params cm.GetRolesParams
		ids    []string
		next   string
		prev   string
		offset bool
	}{
		{"first", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2)}, []string{"a", "b"}, "/spaces/space/roles?pageNext=Mg&limit=2", "", false},
		{"next", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2), PageNext: cm.NewOptString("Mg")}, []string{"c"}, "", "/spaces/space/roles?pagePrev=MA&limit=2", false},
		{"previous", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2), PagePrev: cm.NewOptString("MA")}, []string{"a", "b"}, "/spaces/space/roles?pageNext=Mg&limit=2", "", false},
		{"maximum offset", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2), Skip: cm.NewOptInt64(math.MaxInt)}, []string{}, "", "", true},
		{"maximum limit", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(math.MaxInt), Skip: cm.NewOptInt64(0)}, []string{"a", "b", "c"}, "", "", true},
		{"offset", cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2), Skip: cm.NewOptInt64(1)}, []string{"b", "c"}, "", "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response, err := client.GetRoles(t.Context(), test.params)
			require.NoError(t, err)

			collection, ok := response.(*cm.RoleCollection)
			require.True(t, ok)

			ids := make([]string, 0, len(collection.Items))
			for _, role := range collection.Items {
				ids = append(ids, role.Sys.ID)
			}

			assert.Equal(t, test.ids, ids)
			assert.Equal(t, test.offset, collection.Total.IsSet())
			assert.Equal(t, !test.offset, collection.Pages.IsSet())
			assert.Equal(t, test.params.Limit.Value, int64(collection.Limit.Value))

			if test.offset {
				assert.Equal(t, test.params.Skip.Value, int64(collection.Skip.Value))
			}

			assert.Equal(t, test.next, collection.Pages.Value.Next.Or(""))
			assert.Equal(t, test.prev, collection.Pages.Value.Prev.Or(""))
		})
	}

	invalid := []cm.GetRolesParams{
		{SpaceID: "space", PageNext: cm.NewOptString("!")},
		{SpaceID: "space", PageNext: cm.NewOptString(base64.RawURLEncoding.EncodeToString([]byte("not-an-integer")))},
		{SpaceID: "space", PageNext: cm.NewOptString(base64.RawURLEncoding.EncodeToString([]byte("9223372036854775808")))},
		{SpaceID: "space", PagePrev: cm.NewOptString("LTE")},
		{SpaceID: "space", Skip: cm.NewOptInt64(-1)},
		{SpaceID: "space", Limit: cm.NewOptInt64(0)},
		{SpaceID: "space", PageNext: cm.NewOptString("Mg"), PagePrev: cm.NewOptString("MA")},
		{SpaceID: "space", PageNext: cm.NewOptString("Mg"), Skip: cm.NewOptInt64(0)},
	}
	// On 32-bit platforms, int64 query values and cursor offsets can exceed
	// the int fields in the response. Keep these cases runnable on those targets.
	if strconv.IntSize == 32 {
		invalid = append(invalid,
			cm.GetRolesParams{SpaceID: "space", Skip: cm.NewOptInt64(2147483648)},
			cm.GetRolesParams{SpaceID: "space", Limit: cm.NewOptInt64(2147483648)},
			cm.GetRolesParams{SpaceID: "space", PageNext: cm.NewOptString("MjE0NzQ4MzY0OA")},
		)
	}

	for index, params := range invalid {
		t.Run("invalid "+strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			response, err := client.GetRoles(t.Context(), params)
			require.NoError(t, err)

			failure, ok := response.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, 400, failure.StatusCode)
		})
	}
}
