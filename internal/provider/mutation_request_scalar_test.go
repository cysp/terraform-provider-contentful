package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMutationRequestsRejectUnresolvedRequiredScalars(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		convert      func(*testing.T) diag.Diagnostics
		expectedPath string
	}{
		"app signing secret unknown value": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&AppSigningSecretModel{Value: types.StringUnknown()}).ToAppSigningSecretRequest(t.Context(), path.Empty())
				assert.Equal(t, cm.AppSigningSecretRequestData{}, request)

				return diags
			},
			expectedPath: "value",
		},
		"app signing secret null value": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&AppSigningSecretModel{Value: types.StringNull()}).ToAppSigningSecretRequest(t.Context(), path.Empty())
				assert.Equal(t, cm.AppSigningSecretRequestData{}, request)

				return diags
			},
			expectedPath: "value",
		},
		"environment unknown name": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&EnvironmentModel{Name: types.StringUnknown()}).ToEnvironmentData(t.Context())
				assert.Equal(t, cm.EnvironmentData{}, request)

				return diags
			},
			expectedPath: "name",
		},
		"environment null name": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&EnvironmentModel{Name: types.StringNull()}).ToEnvironmentData(t.Context())
				assert.Equal(t, cm.EnvironmentData{}, request)

				return diags
			},
			expectedPath: "name",
		},
		"environment alias unknown target": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&EnvironmentAliasModel{TargetEnvironmentID: types.StringUnknown()}).ToEnvironmentAliasData(t.Context())
				assert.Equal(t, cm.EnvironmentAliasData{}, request)

				return diags
			},
			expectedPath: "target_environment_id",
		},
		"environment alias null target": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				request, diags := (&EnvironmentAliasModel{TargetEnvironmentID: types.StringNull()}).ToEnvironmentAliasData(t.Context())
				assert.Equal(t, cm.EnvironmentAliasData{}, request)

				return diags
			},
			expectedPath: "target_environment_id",
		},
		"resource provider unknown id": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				model := ResourceProviderModel{ResourceProviderID: types.StringUnknown(), FunctionID: types.StringValue("function")}
				request, diags := model.ToResourceProviderRequest(t.Context(), path.Empty())
				assert.Equal(t, cm.ResourceProviderRequest{}, request)

				return diags
			},
			expectedPath: "resource_provider_id",
		},
		"resource provider null function": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				model := ResourceProviderModel{ResourceProviderID: types.StringValue("provider"), FunctionID: types.StringNull()}
				request, diags := model.ToResourceProviderRequest(t.Context(), path.Empty())
				assert.Equal(t, cm.ResourceProviderRequest{}, request)

				return diags
			},
			expectedPath: "function_id",
		},
		"team space membership unknown admin": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				model := TeamSpaceMembershipModel{Admin: types.BoolUnknown(), Roles: []types.String{}}
				request, diags := model.ToTeamSpaceMembershipData(t.Context(), path.Empty())
				assert.Equal(t, cm.TeamSpaceMembershipData{}, request)

				return diags
			},
			expectedPath: "admin",
		},
		"team space membership null admin": {
			convert: func(t *testing.T) diag.Diagnostics {
				t.Helper()

				model := TeamSpaceMembershipModel{Admin: types.BoolNull(), Roles: []types.String{}}
				request, diags := model.ToTeamSpaceMembershipData(t.Context(), path.Empty())
				assert.Equal(t, cm.TeamSpaceMembershipData{}, request)

				return diags
			},
			expectedPath: "admin",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := test.convert(t)

			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestDeliveryAPIKeyRequestRejectsUnknownScalarsAndPreservesNullPolicy(t *testing.T) {
	t.Parallel()

	model := DeliveryAPIKeyModel{
		Name:         types.StringUnknown(),
		Description:  types.StringUnknown(),
		Environments: NewTypedListNull[types.String](),
	}

	request, diags := model.ToDeliveryAPIKeyRequestData(t.Context(), DeliveryAPIKeyModel{Environments: model.Environments})

	assert.Equal(t, cm.ApiKeyRequestData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"name", "description"}, attributeDiagnosticPaths(t, diags))

	model.Name = types.StringValue("")
	model.Description = types.StringNull()
	request, diags = model.ToDeliveryAPIKeyRequestData(t.Context(), DeliveryAPIKeyModel{Environments: model.Environments})

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, request.Name)
	assert.True(t, request.Description.IsSet())
	assert.True(t, request.Description.IsNull())
}

func TestEnvironmentSourceHeaderRejectsUnknownAndPreservesOmissionPolicy(t *testing.T) {
	t.Parallel()

	model := EnvironmentModel{SourceEnvironmentID: types.StringUnknown()}

	header, diags := model.ToSourceEnvironmentHeader()

	assert.False(t, header.IsSet())
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"source_environment_id"}, attributeDiagnosticPaths(t, diags))

	for name, value := range map[string]types.String{
		"null":  types.StringNull(),
		"empty": types.StringValue(""),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			header, diags := (&EnvironmentModel{SourceEnvironmentID: value}).ToSourceEnvironmentHeader()

			require.False(t, diags.HasError(), diags.Errors())
			assert.False(t, header.IsSet())
		})
	}
}

func TestPersonalAccessTokenRequestRejectsUnknownScalarsAndPreservesNullPolicy(t *testing.T) {
	t.Parallel()

	model := PersonalAccessTokenModel{
		Name:      types.StringUnknown(),
		ExpiresIn: types.Int64Unknown(),
		Scopes:    NewTypedList([]types.String{types.StringValue("scope")}),
	}

	request, diags := model.ToPersonalAccessTokenRequestData(t.Context())

	assert.Equal(t, cm.PersonalAccessTokenRequestData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"name", "expires_in"}, attributeDiagnosticPaths(t, diags))

	model.Name = types.StringValue("")
	model.ExpiresIn = types.Int64Null()
	request, diags = model.ToPersonalAccessTokenRequestData(t.Context())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, request.Name)
	assert.False(t, request.ExpiresIn.IsSet())
}

func TestResourceTypeRequestRejectsUnresolvedNestedScalarsAtomically(t *testing.T) {
	t.Parallel()

	model := validResourceTypeRequestModel()
	model.Name = types.StringUnknown()
	model.DefaultFieldMapping.Title = types.StringNull()
	model.DefaultFieldMapping.Subtitle = types.StringUnknown()
	model.DefaultFieldMapping.Image.URL = types.StringUnknown()
	model.DefaultFieldMapping.Image.AltText = types.StringUnknown()
	model.DefaultFieldMapping.Badge.Label = types.StringNull()
	model.DefaultFieldMapping.Badge.Variant = types.StringUnknown()

	request, diags := model.ToResourceTypeData(t.Context(), path.Empty())

	assert.Equal(t, cm.ResourceTypeData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		"name",
		"default_field_mapping.title",
		"default_field_mapping.subtitle",
		"default_field_mapping.image.url",
		"default_field_mapping.image.alt_text",
		"default_field_mapping.badge.label",
		"default_field_mapping.badge.variant",
	}, attributeDiagnosticPaths(t, diags))
}

func TestResourceTypeRequestRejectsMissingRequiredMapping(t *testing.T) {
	t.Parallel()

	model := validResourceTypeRequestModel()
	model.DefaultFieldMapping = nil

	request, diags := model.ToResourceTypeData(t.Context(), path.Empty())

	assert.Equal(t, cm.ResourceTypeData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"default_field_mapping"}, attributeDiagnosticPaths(t, diags))
}

func TestResourceTypeRequestPreservesOptionalNullAndKnownEmptyValues(t *testing.T) {
	t.Parallel()

	model := validResourceTypeRequestModel()
	model.Name = types.StringValue("")
	model.DefaultFieldMapping.Title = types.StringValue("")
	model.DefaultFieldMapping.Subtitle = types.StringNull()
	model.DefaultFieldMapping.Description = types.StringValue("")
	model.DefaultFieldMapping.ExternalURL = types.StringNull()
	model.DefaultFieldMapping.Image.AltText = types.StringValue("")

	request, diags := model.ToResourceTypeData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, request.Name)
	assert.Empty(t, request.DefaultFieldMapping.Title)
	assert.False(t, request.DefaultFieldMapping.Subtitle.IsSet())
	description, ok := request.DefaultFieldMapping.Description.Get()
	require.True(t, ok)
	assert.Empty(t, description)
	assert.False(t, request.DefaultFieldMapping.ExternalUrl.IsSet())
	image, ok := request.DefaultFieldMapping.Image.Get()
	require.True(t, ok)
	altText, ok := image.AltText.Get()
	require.True(t, ok)
	assert.Empty(t, altText)
}

func TestTagPutRequestRejectsUnresolvedScalarsAtomically(t *testing.T) {
	t.Parallel()

	model := validTagRequestModel()
	model.SpaceID = types.StringUnknown()
	model.EnvironmentID = types.StringNull()
	model.TagID = types.StringUnknown()
	model.Name = types.StringNull()
	model.Visibility = types.StringUnknown()

	params, request, diags := model.ToPutTagRequest(path.Empty())

	assert.Equal(t, cm.PutTagParams{}, params)
	assert.Equal(t, cm.TagRequest{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"space_id", "environment_id", "tag_id", "name", "visibility"}, attributeDiagnosticPaths(t, diags))
}

func validResourceTypeRequestModel() ResourceTypeModel {
	return ResourceTypeModel{
		Name: types.StringValue("resource type"),
		DefaultFieldMapping: &ResourceTypeFieldMapping{
			Title:       types.StringValue("title"),
			Subtitle:    types.StringNull(),
			Description: types.StringNull(),
			ExternalURL: types.StringNull(),
			Image: &ResourceTypeFieldMappingImage{
				URL:     types.StringValue("image"),
				AltText: types.StringNull(),
			},
			Badge: &ResourceTypeFieldMappingBadge{
				Label:   types.StringValue("label"),
				Variant: types.StringValue("primary"),
			},
		},
	}
}

func validTagRequestModel() TagModel {
	return TagModel{
		TagIdentityModel: TagIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			TagID:         types.StringValue("tag"),
		},
		Name:       types.StringValue("Tag"),
		Visibility: types.StringValue("private"),
	}
}
