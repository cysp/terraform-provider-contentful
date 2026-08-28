package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedIgnoredDriftContentType = errors.New("unexpected final ignored-drift Content Type")

func TestAccContentTypeResourceIgnoreChangesDoesNotAuthorizeExternalDraft(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	contentTypeID := "ignore-changes-activation-authority"

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: contentTypeIgnoreChangesConfig(contentTypeID, "Managed one")},
		{
			PreConfig: func() {
				currentResponse, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
					SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
				})
				require.NoError(t, getErr)

				current, ok := currentResponse.(*cm.ContentType)
				require.True(t, ok)

				wireCurrent, marshalErr := json.Marshal(current)
				require.NoError(t, marshalErr)

				var externalDraft cm.ContentTypeRequestData

				require.NoError(t, json.Unmarshal(wireCurrent, &externalDraft))
				externalDraft.Description = cm.NewOptNilString("External ignored description")

				putResponse, putErr := server.Handler().PutContentType(t.Context(), &externalDraft, cm.PutContentTypeParams{
					SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
					XContentfulVersion: cm.NewOptInt(current.Sys.Version),
				})
				require.NoError(t, putErr)

				draft, ok := putResponse.(*cm.ContentTypeStatusCode)
				require.True(t, ok)
				require.Equal(t, 3, draft.Response.Sys.Version)
				require.Equal(t, 1, draft.Response.Sys.PublishedVersion.Or(0))
				handler.resetRequestHistory()
			},
			Config: contentTypeIgnoreChangesConfig(contentTypeID, "Managed one"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("description"), knownvalue.StringExact("External ignored description")),
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
			},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
		{
			PreConfig: handler.resetRequestHistory,
			Config:    contentTypeIgnoreChangesConfig(contentTypeID, "Managed two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("name"), knownvalue.StringExact("Managed two")),
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("description"), knownvalue.StringExact("External ignored description")),
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(4)),
			},
			Check: func(state *terraform.State) error {
				checkErr := contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{4})(state)
				if checkErr != nil {
					return checkErr
				}

				body := handler.lastPutBody()

				var requestBody map[string]any

				require.NoError(t, json.Unmarshal(body, &requestBody))
				require.Equal(t, "Managed two", requestBody["name"])
				require.Equal(t, "External ignored description", requestBody["description"])

				response, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
					SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
				})
				if getErr != nil {
					return fmt.Errorf("read final ignored-drift Content Type: %w", getErr)
				}

				contentType, ok := response.(*cm.ContentType)
				if !ok {
					return fmt.Errorf("%w: got %T", errUnexpectedContentTypeResponseType, response)
				}

				if contentType.Name != "Managed two" || contentType.Description.Or("") != "External ignored description" ||
					contentType.Sys.Version != 5 || contentType.Sys.PublishedVersion.Or(0) != 4 {
					return fmt.Errorf("%w: name=%q description=%q version=%d published=%d", errUnexpectedIgnoredDriftContentType,
						contentType.Name, contentType.Description.Or(""), contentType.Sys.Version, contentType.Sys.PublishedVersion.Or(0))
				}

				return nil
			},
		},
	}})
}

func contentTypeIgnoreChangesConfig(contentTypeID, name string) string {
	return fmt.Sprintf(`
resource "contentful_content_type" "test" {
  space_id        = "space"
  environment_id  = "environment"
  content_type_id = %q
  name             = %q
  description      = "Managed description"
  display_field    = "name"

  fields = [{
    id        = "name"
    name      = "Name"
    type      = "Symbol"
    required  = true
    localized = false
  }]

  lifecycle {
    ignore_changes = [description]
  }
}
`, contentTypeID, name)
}
