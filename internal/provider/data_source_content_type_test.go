package provider_test

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeDataSourcesComposition(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetContentType("space", "master", "article", cm.ContentTypeRequestData{
		Name:         "Article",
		DisplayField: "title",
		Fields: []cm.ContentTypeRequestDataFieldsItem{{
			ID:       "title",
			Name:     "Title",
			Type:     "Symbol",
			Required: cm.NewOptBool(true),
		}},
	})

	var contentTypeReads, contentTypeMutations atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/content_types") {
			if r.Method == http.MethodGet {
				contentTypeReads.Add(1)
			} else {
				contentTypeMutations.Add(1)
			}
		}

		server.ServeHTTP(w, r)
	})

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectUnknownValue("data.contentful_content_types.all", tfjsonpath.New("content_types")),
				plancheck.ExpectUnknownValue("data.contentful_content_type.article", tfjsonpath.New("name")),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("data.contentful_content_types.all", tfjsonpath.New("id"), knownvalue.StringExact("space/master")),
				statecheck.ExpectKnownValue("data.contentful_content_types.all", tfjsonpath.New("content_types"), knownvalue.ListSizeExact(1)),
				statecheck.ExpectKnownValue("data.contentful_content_types.all", tfjsonpath.New("content_types").AtSliceIndex(0).AtMapKey("content_type_id"), knownvalue.StringExact("article")),
				statecheck.ExpectKnownValue("data.contentful_content_type.article", tfjsonpath.New("id"), knownvalue.StringExact("space/master/article")),
				statecheck.ExpectKnownValue("data.contentful_content_type.article", tfjsonpath.New("name"), knownvalue.StringExact("Article")),
				statecheck.ExpectKnownValue("contentful_entry.using_lookup", tfjsonpath.New("content_type_id"), knownvalue.StringExact("article")),
			},
		},
		{ConfigDirectory: config.TestNameDirectory(), PlanOnly: true},
	}})

	assert.Positive(t, contentTypeReads.Load())
	assert.Zero(t, contentTypeMutations.Load())
}
