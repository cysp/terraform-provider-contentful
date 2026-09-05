package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccEditorInterfaceResourceCreateAfterFirstActivationRecovery(t *testing.T) {
	t.Parallel()

	for _, noRefresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("no_refresh=%t", noRefresh), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			activation := &contentTypeActivationTestHandler{delegate: server}
			activation.failActivation.Store(true)
			handler := &editorInterfaceRequestRecorder{next: activation}
			contentType := editorInterfaceActivationContentTypeConfig("initial")

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: noRefresh}},
				Steps: []resource.TestStep{
					{Config: contentType, ExpectNonEmptyPlan: true},
					{
						PreConfig: func() { activation.failActivation.Store(false) },
						Config:    contentType + editorInterfaceActivationConfig("initial"),
					},
					{Config: contentType + editorInterfaceActivationConfig("updated")},
				},
			})

			// First activation creates version 1, even when Terraform recovers via Update.
			// The next PUT must use the version returned by the first interface mutation.
			require.Equal(t, []string{"PUT:1", "PUT:2"}, slices.DeleteFunc(handler.Requests(), func(request string) bool { return request == http.MethodGet }))
		})
	}
}

func TestAccEditorInterfaceResourceCreateAfterImportedDraftActivation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetContentType("space", "environment", "editor-offset", cm.ContentTypeRequestData{
		Name: "Offset", Description: cm.NewOptNilString("initial"), DisplayField: "name",
		Fields: []cm.ContentTypeRequestDataFieldsItem{{ID: "name", Name: "Name", Type: "Symbol"}},
	})
	handler := &editorInterfaceRequestRecorder{next: server}

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:       editorInterfaceActivationContentTypeConfig("initial"),
				ResourceName: "contentful_content_type.test", ImportState: true,
				ImportStateId: "space/environment/editor-offset", ImportStatePersist: true,
			},
			{Config: editorInterfaceActivationContentTypeConfig("updated") + editorInterfaceActivationConfig("initial")},
		},
	})

	require.Equal(t, []string{"PUT:1", "GET"}, handler.Requests())
}

func TestAccEditorInterfaceResourceUpdateAfterActivationRecovery(t *testing.T) {
	t.Parallel()

	for _, noRefresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("no_refresh=%t", noRefresh), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			activation := &contentTypeActivationTestHandler{delegate: server}
			handler := &editorInterfaceRequestRecorder{next: activation}
			updated := editorInterfaceActivationContentTypeConfig("updated")

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: noRefresh}},
				Steps: []resource.TestStep{
					{Config: editorInterfaceActivationContentTypeConfig("initial") + editorInterfaceActivationConfig("initial")},
					{
						PreConfig:   func() { activation.failActivation.Store(true) },
						Config:      updated + editorInterfaceActivationConfig("updated"),
						ExpectError: regexp.MustCompile(`Failed to activate content type`),
					},
					{
						PreConfig: func() { activation.failActivation.Store(false) },
						Config:    updated + editorInterfaceActivationConfig("updated"),
					},
					{Config: updated + editorInterfaceActivationConfig("third")},
				},
			})

			require.Equal(t, []string{"PUT:1", "PUT:3", "PUT:4"}, slices.DeleteFunc(handler.Requests(), func(request string) bool { return request == http.MethodGet }))
		})
	}
}

func editorInterfaceActivationContentTypeConfig(description string) string {
	return fmt.Sprintf(`
resource "contentful_content_type" "test" {
 space_id = "space"
 environment_id = "environment"
 content_type_id = "editor-offset"
 name = "Offset"
 description = %[1]q
 display_field = "name"
 fields = [{id="name",name="Name",type="Symbol",required=false,localized=false}]
}
`, description)
}

func editorInterfaceActivationConfig(helpText string) string {
	return fmt.Sprintf(`
resource "contentful_editor_interface" "test" {
 space_id = contentful_content_type.test.space_id
 environment_id = contentful_content_type.test.environment_id
 content_type_id = contentful_content_type.test.content_type_id
 controls = [{
  field_id = "name"
  widget_id = "singleLine"
  widget_namespace = "builtin"
  settings = jsonencode({helpText = %[1]q})
 }]
}
`, helpText)
}
