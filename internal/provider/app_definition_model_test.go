package provider_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func FuzzAppDefinitionResourceModelRoundtrip(f *testing.F) {
	corpus := []cm.AppDefinition{
		{
			Sys:  cm.AppDefinitionSys{Type: "AppDefinition", Organization: cm.OrganizationLink{Sys: cm.OrganizationLinkSys{Type: "Link", LinkType: "Organization", ID: "organization-id"}}, ID: "app-definition-id"},
			Name: "App Definition Name",
		},
		{
			Sys:  cm.AppDefinitionSys{Type: "AppDefinition", Organization: cm.OrganizationLink{Sys: cm.OrganizationLinkSys{Type: "Link", LinkType: "Organization", ID: "organization-id"}}, ID: "app-definition-id"},
			Name: "App Definition Name",
			Src:  cm.NewOptString("https://example.com/app-definition.js"),
			Bundle: cm.NewOptAppBundleLink(cm.AppBundleLink{
				Sys: cm.AppBundleLinkSys{Type: cm.AppBundleLinkSysTypeLink, LinkType: cm.AppBundleLinkSysLinkTypeAppBundle, ID: "app-bundle-id"},
			}),
			Locations: []cm.AppDefinitionLocationsItem{
				{
					Location: "app-config",
				},
				{
					Location: "entry-field",
					FieldTypes: []cm.AppDefinitionLocationsItemFieldTypesItem{
						{
							Type: "Symbol",
						},
						{
							Type: "Link", LinkType: cm.NewOptString("Entry"),
						},
						{
							Type: "Array",
							Items: cm.NewOptAppDefinitionLocationsItemFieldTypesItemItems(
								cm.AppDefinitionLocationsItemFieldTypesItemItems{Type: "Symbol"},
							),
						},
						{
							Type: "Array",
							Items: cm.NewOptAppDefinitionLocationsItemFieldTypesItemItems(
								cm.AppDefinitionLocationsItemFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Entry")},
							),
						},
					},
				},
				{
					Location: "page",
					NavigationItem: cm.NewOptAppDefinitionLocationsItemNavigationItem(cm.AppDefinitionLocationsItemNavigationItem{
						Name: "Page",
						Path: "/page",
					}),
				},
			},
			Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{
				Installation: []cm.AppDefinitionParameter{
					{
						ID: "parameter-a",
						Labels: cm.NewOptAppDefinitionParameterLabels(cm.AppDefinitionParameterLabels{
							Empty: cm.NewOptString("empty"),
						}),
					},
				},
				Instance: []cm.AppDefinitionParameter{
					{
						ID:      "parameter-b",
						Options: []jx.Raw{[]byte(`"option-a"`)},
					},
				},
			}),
		},
	}

	for _, appDefinition := range corpus {
		appDefinitionJSON, appDefinitionJSONErr := json.Marshal(&appDefinition)
		if appDefinitionJSONErr == nil {
			f.Add(appDefinitionJSON)
		} else {
			f.Fatal(appDefinitionJSONErr)
		}
	}

	f.Fuzz(func(t *testing.T, inputBytes []byte) {
		var input cm.AppDefinition

		err := json.Unmarshal(inputBytes, &input)
		if err != nil {
			t.Skipf("Skipping invalid JSON: %v", err)
		}

		input.Sys.Type = cm.AppDefinitionSysTypeAppDefinition
		input.Sys.Organization.Sys.Type = cm.OrganizationLinkSysTypeLink
		input.Sys.Organization.Sys.LinkType = cm.OrganizationLinkSysLinkTypeOrganization

		if input.Bundle.IsSet() {
			input.Bundle.Value.Sys.Type = cm.AppBundleLinkSysTypeLink
			input.Bundle.Value.Sys.LinkType = cm.AppBundleLinkSysLinkTypeAppBundle
		}

		model, modelDiags := provider.NewAppDefinitionResourceModelFromResponse(t.Context(), input)
		if modelDiags.HasError() {
			t.Fatalf("Failed to convert AppDefinition to AppDefinitionResourceModel: %v", modelDiags)
		}

		appDefinitionFields, appDefinitionFieldsDiags := model.ToAppDefinitionFields(t.Context(), path.Empty())
		if appDefinitionFieldsDiags.HasError() {
			t.Fatalf("Failed to convert AppDefinitionResourceModel to AppDefinitionFields: %v", appDefinitionFieldsDiags)
		}

		output := cmts.NewAppDefinitionFromFields(input.Sys.Organization.Sys.ID, input.Sys.ID, appDefinitionFields)

		assert.Equal(t, input, output, "AppDefinition should be equal after roundtrip conversion")
	})
}
