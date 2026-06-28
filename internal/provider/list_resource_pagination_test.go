//nolint:testpackage
package provider

import (
	"context"
	"errors"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTestContentfulUnavailable = errors.New("contentful unavailable")

type testListCollection struct {
	total cm.OptInt
	items []int
}

func (c testListCollection) GetTotal() cm.OptInt {
	return c.total
}

func (c testListCollection) GetItems() []int {
	return c.items
}

type testListParams struct {
	skip  int64
	limit int64
}

func TestPaginateContentfulCollectionItemsAsListResultsPaginatesUntilTotal(t *testing.T) {
	t.Parallel()

	var requests []testListParams

	results := collectTestListResults(t, paginateContentfulCollectionItemsAsListResults(
		context.Background(),
		list.ListRequest{},
		"failed",
		func(_ context.Context, skip int64, limit int64) (testListCollection, error) {
			params := testListParams{skip: skip, limit: limit}
			requests = append(requests, params)

			switch params.skip {
			case 0:
				return testListCollection{total: cm.NewOptInt(3), items: []int{1, 2}}, nil
			case 2:
				return testListCollection{total: cm.NewOptInt(3), items: []int{3}}, nil
			default:
				t.Fatalf("unexpected skip %d", params.skip)

				return testListCollection{}, nil
			}
		},
		testListResult,
	))

	assert.Equal(t, []testListParams{
		{skip: 0, limit: defaultPageLimit},
		{skip: 2, limit: defaultPageLimit},
	}, requests)
	assert.Equal(t, []string{"1", "2", "3"}, results)
}

func TestPaginateContentfulCollectionItemsAsListResultsHonorsTerraformLimit(t *testing.T) {
	t.Parallel()

	var requests []testListParams

	results := collectTestListResults(t, paginateContentfulCollectionItemsAsListResults(
		context.Background(),
		list.ListRequest{Limit: 2},
		"failed",
		func(_ context.Context, skip int64, limit int64) (testListCollection, error) {
			requests = append(requests, testListParams{skip: skip, limit: limit})

			return testListCollection{total: cm.NewOptInt(10), items: []int{1, 2}}, nil
		},
		testListResult,
	))

	assert.Equal(t, []testListParams{{skip: 0, limit: 2}}, requests)
	assert.Equal(t, []string{"1", "2"}, results)
}

func TestPaginateContentfulCollectionItemsAsListResultsReturnsFetchDiagnostics(t *testing.T) {
	t.Parallel()

	results := collectTestListResults(t, paginateContentfulCollectionItemsAsListResults(
		context.Background(),
		list.ListRequest{},
		"failed",
		func(context.Context, int64, int64) (testListCollection, error) {
			return testListCollection{}, errTestContentfulUnavailable
		},
		testListResult,
	))

	require.Len(t, results, 1)
	assert.Equal(t, "diagnostic: contentful unavailable", results[0])
}

func TestPaginateContentfulCollectionItemsAsListResultsReturnsUnexpectedResponseDiagnostics(t *testing.T) {
	t.Parallel()

	results := collectTestListResults(t, paginateContentfulCollectionItemsAsListResults(
		context.Background(),
		list.ListRequest{},
		"failed",
		func(context.Context, int64, int64) (any, error) {
			return struct{}{}, nil
		},
		testListResult,
	))

	require.Len(t, results, 1)
	assert.Contains(t, results[0], "diagnostic: Unexpected response type")
}

func TestPaginateContentfulCollectionItemsAsListResultsReturnsContentfulErrorDiagnostics(t *testing.T) {
	t.Parallel()

	results := collectTestListResults(t, paginateContentfulCollectionItemsAsListResults(
		context.Background(),
		list.ListRequest{},
		"failed",
		func(context.Context, int64, int64) (any, error) {
			return &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("NotFound"),
					Message: cm.NewOptString("Environment not found"),
				}),
			}, nil
		},
		testListResult,
	))

	require.Len(t, results, 1)
	assert.Equal(t, "diagnostic: Error: NotFound: Environment not found", results[0])
}

func collectTestListResults(t *testing.T, stream func(func(list.ListResult) bool)) []string {
	t.Helper()

	var results []string

	stream(func(result list.ListResult) bool {
		if result.Diagnostics.HasError() {
			results = append(results, "diagnostic: "+result.Diagnostics.Errors()[0].Detail())

			return true
		}

		results = append(results, result.DisplayName)

		return true
	})

	return results
}

func testListResult(item int) list.ListResult {
	return list.ListResult{DisplayName: strconv.Itoa(item)}
}
