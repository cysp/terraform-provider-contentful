package provider_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func FuzzAppDefinitionResourceModelRoundTrip(f *testing.F) {
	corpus := []cm.AppDefinition{
		{
			Sys:  cm.NewAppDefinitionSys("organization-id", "app-definition-id"),
			Name: "App Definition Name",
		},
		{
			Sys:    cm.NewAppDefinitionSys("organization-id", "app-definition-id"),
			Name:   "App Definition Name",
			Src:    cm.NewOptString("https://example.com/app-definition.js"),
			Bundle: cm.NewOptAppBundleLink(cm.NewAppBundleLink("app-bundle-id")),
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
						Options: []jx.Raw{[]byte(testJSON("option-a"))},
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

	seed := []byte(`{"name":"App Definition Name",
		"parameters":{"installation":[{"id":"parameter-a","type":"Symbol","name":"Parameter A","default":{ "b": [null, {}, [], 9007199254740993, 1.234567890123456789, -0, 1e+30], "a": 1 },"options":[{ "b":2, "a":1 },null,{},[]]}],"instance":[]},
		"sys":{"id":"app-definition-id","type":"AppDefinition","organization":{"sys":{"type":"Link","linkType":"Organization","id":"organization-id"}}}}`)

	var decodedSeed cm.AppDefinition

	err := json.Unmarshal(seed, &decodedSeed)
	if err != nil {
		f.Fatal(err)
	}

	f.Add(seed)

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

		model, modelDiags := NewAppDefinitionResourceModelFromResponse(t.Context(), input)
		if modelDiags.HasError() {
			t.Fatalf("Failed to convert AppDefinition to AppDefinitionResourceModel: %v", modelDiags)
		}

		appDefinitionFields, appDefinitionFieldsDiags := model.ToAppDefinitionData(AppDefinitionBaseModel{}, path.Empty())
		if appDefinitionFieldsDiags.HasError() {
			t.Fatalf("Failed to convert AppDefinitionResourceModel to AppDefinitionData: %v", appDefinitionFieldsDiags)
		}

		output := cmt.NewAppDefinitionFromFields(input.Sys.Organization.Sys.ID, input.Sys.ID, appDefinitionFields)
		input.Parameters = canonicalAppDefinitionParameters(t, input.Parameters)
		output.Parameters = canonicalAppDefinitionParameters(t, output.Parameters)

		assert.Equal(t, input, output, "AppDefinition should be equal after roundtrip conversion")
	})
}
