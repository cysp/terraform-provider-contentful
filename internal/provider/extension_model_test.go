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

func FuzzExtensionModelRoundTrip(f *testing.F) {
	corpus := []cm.Extension{
		{
			Sys: cm.NewExtensionSys("space-id", "environment-id", "extension-id"),
			Extension: cm.ExtensionExtension{
				Name:   "Extension Name",
				Srcdoc: cm.NewOptString(""),
			},
		},
		{
			Sys: cm.NewExtensionSys("space-id", "environment-id", "extension-id"),
			Extension: cm.ExtensionExtension{
				Name: "Extension Name",
				Src:  cm.NewOptString("https://example.com/extension.js"),
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
							Options: []jx.Raw{[]byte(testJSON("option-a"))},
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

	// Keep representation differences in the seed: struct marshaling compacts
	// embedded JSON before the target sees it.
	seed := []byte(`{"extension":{"name":"Extension Name","srcdoc":"","fieldTypes":[]},
		"parameters":{ "b": [null, {}, [], 9007199254740993, 1.234567890123456789, -0, 1e+30], "a": 1 },
		"sys":{"id":"extension-id","type":"Extension","version":1,"space":{"sys":{"type":"Link","linkType":"Space","id":"space-id"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"environment-id"}}}}`)

	var decodedSeed cm.Extension

	err := json.Unmarshal(seed, &decodedSeed)
	if err != nil {
		f.Fatal(err)
	}

	f.Add(seed)

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

		_, srcSet := input.Extension.Src.Get()
		_, srcdocSet := input.Extension.Srcdoc.Get()

		if srcSet == srcdocSet {
			t.Skip("Skipping response that violates the Contentful extension source union")
		}

		model, modelDiags := NewExtensionModelFromResponse(t.Context(), input)
		if modelDiags.HasError() {
			t.Fatalf("Failed to convert Extension to ExtensionModel: %v", modelDiags)
		}

		extensionFields, extensionFieldsDiags := model.ToExtensionData(model, path.Empty())
		if extensionFieldsDiags.HasError() {
			t.Fatalf("Failed to convert ExtensionModel to ExtensionData: %v", extensionFieldsDiags)
		}

		output := cmt.NewExtensionFromFields(input.Sys.Space.Sys.ID, input.Sys.Environment.Sys.ID, input.Sys.ID, extensionFields)
		// Version belongs to resource private state, not the configuration model.
		output.Sys.Version = input.Sys.Version
		input.Parameters = canonicalModelJSON(t, input.Parameters)
		output.Parameters = canonicalModelJSON(t, output.Parameters)
		input.Extension.Parameters = canonicalAppDefinitionParameters(t, input.Extension.Parameters)
		output.Extension.Parameters = canonicalAppDefinitionParameters(t, output.Extension.Parameters)

		assert.Equal(t, input, output, "Extension should be equal after roundtrip conversion")
	})
}
