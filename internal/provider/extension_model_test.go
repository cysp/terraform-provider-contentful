package provider_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func FuzzExtensionModelRoundtrip(f *testing.F) {
	corpus := []cm.Extension{
		{
			Sys: cm.ExtensionSys{Type: "Extension", Space: cm.SpaceLink{Sys: cm.SpaceLinkSys{Type: "Link", LinkType: "Space", ID: "space-id"}}, Environment: cm.EnvironmentLink{Sys: cm.EnvironmentLinkSys{Type: "Link", LinkType: "Environment", ID: "environment-id"}}, ID: "extension-id"},
			Extension: cm.ExtensionExtension{
				Name: "Extension Name",
			},
		},
		{
			Sys: cm.ExtensionSys{Type: "Extension", Space: cm.SpaceLink{Sys: cm.SpaceLinkSys{Type: "Link", LinkType: "Space", ID: "space-id"}}, Environment: cm.EnvironmentLink{Sys: cm.EnvironmentLinkSys{Type: "Link", LinkType: "Environment", ID: "environment-id"}}, ID: "extension-id"},
			Extension: cm.ExtensionExtension{
				Name:   "Extension Name",
				Src:    cm.NewOptString("https://example.com/extension.js"),
				Srcdoc: cm.NewOptString("<!DOCTYPE html><html><head><title>Extension</title></head><body><h1>Extension</h1></body></html>"),
				FieldTypes: []cm.ExtensionExtensionFieldTypesItem{
					{
						Type: "Symbol",
					},
					{
						Type: "Link", LinkType: cm.NewOptString("Entry"),
					},
					{
						Type: "Array",
						Items: cm.NewOptExtensionExtensionFieldTypesItemItems(
							cm.ExtensionExtensionFieldTypesItemItems{Type: "Symbol"},
						),
					},
					{
						Type: "Array",
						Items: cm.NewOptExtensionExtensionFieldTypesItemItems(
							cm.ExtensionExtensionFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Entry")},
						),
					},
				},
				Sidebar: cm.NewOptBool(true),
				Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{
					Installation: []cm.AppDefinitionParameter{
						{
							ID: "parameter-a",
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
		},
	}

	for _, extension := range corpus {
		extensionJSON, extensionJSONErr := json.Marshal(&extension)
		if extensionJSONErr == nil {
			f.Add(extensionJSON)
		} else {
			f.Fatal(extensionJSONErr)
		}
	}

	f.Fuzz(func(t *testing.T, inputBytes []byte) {
		var input cm.Extension

		err := json.Unmarshal(inputBytes, &input)
		if err != nil {
			t.Skipf("Skipping invalid JSON: %v", err)
		}

		input.Sys.Type = cm.ExtensionSysTypeExtension
		input.Sys.Space.Sys.Type = cm.SpaceLinkSysTypeLink
		input.Sys.Space.Sys.LinkType = cm.SpaceLinkSysLinkTypeSpace
		input.Sys.Environment.Sys.Type = cm.EnvironmentLinkSysTypeLink
		input.Sys.Environment.Sys.LinkType = cm.EnvironmentLinkSysLinkTypeEnvironment

		if input.Extension.Src.Value == "" {
			input.Extension.Src.Reset()
		}

		if input.Extension.Srcdoc.Value == "" {
			input.Extension.Srcdoc.Reset()
		}

		model, modelDiags := provider.NewExtensionModelFromResponse(t.Context(), input)
		if modelDiags.HasError() {
			t.Fatalf("Failed to convert Extension to ExtensionModel: %v", modelDiags)
		}

		extensionFields, extensionFieldsDiags := model.ToExtensionFields(t.Context(), path.Empty())
		if extensionFieldsDiags.HasError() {
			t.Fatalf("Failed to convert ExtensionModel to ExtensionFields: %v", extensionFieldsDiags)
		}

		output := cmt.NewExtensionFromFields(input.Sys.Space.Sys.ID, input.Sys.Environment.Sys.ID, input.Sys.ID, extensionFields)

		assert.Equal(t, input, output, "Extension should be equal after roundtrip conversion")
	})
}
