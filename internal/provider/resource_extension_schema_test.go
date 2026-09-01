package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestExtensionSourceSchemaMatchesContentfulContract(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		src        types.String
		srcdoc     types.String
		wantErrors bool
	}{
		"neither source is configured": {
			src:    types.StringNull(),
			srcdoc: types.StringNull(),
		},
		"empty src is invalid": {
			src:        types.StringValue(""),
			srcdoc:     types.StringNull(),
			wantErrors: true,
		},
		"empty srcdoc is an explicit valid source": {
			src:    types.StringNull(),
			srcdoc: types.StringValue(""),
		},
		"non-empty src is valid": {
			src:    types.StringValue("https://example.com/extension.js"),
			srcdoc: types.StringNull(),
		},
		"both sources are invalid": {
			src:        types.StringValue("https://example.com/extension.js"),
			srcdoc:     types.StringValue("<!doctype html>"),
			wantErrors: true,
		},
		"unknown expression defers validation until planning resolves it": {
			src:    types.StringNull(),
			srcdoc: types.StringUnknown(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			resourceSchema := ExtensionResourceSchema(ctx)
			model := ExtensionModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringNull()},
				ExtensionIdentityModel: ExtensionIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					ExtensionID:   types.StringValue("extension"),
				},
				Extension: &ExtensionConfiguration{
					Name:       types.StringValue("Extension"),
					Src:        test.src,
					SrcDoc:     test.srcdoc,
					FieldTypes: []AppDefinitionLocationFieldTypesItem{},
					Sidebar:    types.BoolValue(false),
				},
				Parameters: NewNormalizedJSONTypesNormalizedValue([]byte(`{}`)),
				Timeouts:   TimeoutsNull(),
			}

			configPlan := tfsdk.Plan{Schema: resourceSchema}
			diags := configPlan.Set(ctx, model)
			require.False(t, diags.HasError(), diags.Errors())

			config := tfsdk.Config{Raw: configPlan.Raw, Schema: resourceSchema}

			srcAttribute, ok := ExtensionResourceExtensionSchemaAttributes(ctx)["src"].(schema.StringAttribute)
			require.True(t, ok)

			request := validator.StringRequest{
				Config:         config,
				ConfigValue:    test.src,
				Path:           path.Root("extension").AtName("src"),
				PathExpression: path.MatchRoot("extension").AtName("src"),
			}

			response := validator.StringResponse{}
			for _, stringValidator := range srcAttribute.Validators {
				stringValidator.ValidateString(ctx, request, &response)
			}

			if test.wantErrors {
				require.True(t, response.Diagnostics.HasError(), response.Diagnostics)
			} else {
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			}
		})
	}
}
