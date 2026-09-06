//nolint:testpackage // Resource lifecycle methods are intentionally tested through their package-local implementations.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:maintidx // One table intentionally covers every hardened Create converter and its diagnostic path.
func TestCreateRequestConversionErrorsStopBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	for name, test := range map[string]struct {
		model          any
		config         any
		resourceSchema schema.Schema
		create         func(*cm.Client, resource.CreateRequest, *resource.CreateResponse)
		expectedPath   string
	}{
		"app signing secret": {
			model: AppSigningSecretModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
				},
				Value:    types.StringUnknown(),
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: AppSigningSecretResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "value",
		},
		"delivery API key": {
			model: DeliveryAPIKeyModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
					SpaceID:  types.StringValue("space"),
					APIKeyID: types.StringUnknown(),
				},
				Name:            types.StringUnknown(),
				Description:     types.StringNull(),
				Environments:    NewTypedListUnknown[types.String](),
				AccessToken:     types.StringUnknown(),
				PreviewAPIKeyID: types.StringUnknown(),
				Timeouts:        TimeoutsNull(),
			},
			config: DeliveryAPIKeyModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringNull()},
				DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
					SpaceID:  types.StringValue("space"),
					APIKeyID: types.StringNull(),
				},
				Name:            types.StringValue("configured key"),
				Description:     types.StringNull(),
				Environments:    NewTypedListNull[types.String](),
				AccessToken:     types.StringNull(),
				PreviewAPIKeyID: types.StringNull(),
				Timeouts:        TimeoutsNull(),
			},
			resourceSchema: DeliveryAPIKeyResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := deliveryAPIKeyResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "name",
		},
		"personal access token": {
			model: PersonalAccessTokenModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				Name:            types.StringUnknown(),
				ExpiresIn:       types.Int64Null(),
				ExpiresAt:       timetypes.NewRFC3339Unknown(),
				RevokedAt:       timetypes.NewRFC3339Unknown(),
				Scopes:          NewTypedList([]types.String{types.StringValue("content_management_read")}),
				Token:           types.StringUnknown(),
				Timeouts:        personalAccessTokenTimeoutsNull(),
			},
			resourceSchema: PersonalAccessTokenResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := personalAccessTokenResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "name",
		},
		"environment": {
			model: EnvironmentModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				EnvironmentIdentityModel: EnvironmentIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
				},
				Name:                types.StringValue("Environment"),
				Status:              types.StringUnknown(),
				SourceEnvironmentID: types.StringUnknown(),
				Timeouts:            TimeoutsNull(),
			},
			resourceSchema: EnvironmentResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := environmentResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "source_environment_id",
		},
		"environment alias": {
			model: EnvironmentAliasModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				EnvironmentAliasIdentityModel: EnvironmentAliasIdentityModel{
					SpaceID:            types.StringValue("space"),
					EnvironmentAliasID: types.StringValue("alias"),
				},
				TargetEnvironmentID: types.StringUnknown(),
				Timeouts:            TimeoutsNull(),
			},
			resourceSchema: EnvironmentAliasResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := environmentAliasResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "target_environment_id",
		},
		"entry": {
			model: EntryModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				EntryIdentityModel: EntryIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					EntryID:       types.StringUnknown(),
				},
				ContentTypeID: types.StringValue("content-type"),
				Fields:        NewTypedMapUnknown[jsontypes.Normalized](),
				Metadata: NewTypedObject(EntryMetadataValue{
					Concepts: NewTypedList([]types.String{}),
					Tags:     NewTypedList([]types.String{}),
				}),
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: EntryResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := entryResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "fields",
		},
		"resource provider": {
			model: ResourceProviderModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				ResourceProviderIdentityModel: ResourceProviderIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
				},
				ResourceProviderID: types.StringValue("provider"),
				FunctionID:         types.StringUnknown(),
				Timeouts:           TimeoutsNull(),
			},
			resourceSchema: ResourceProviderResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := appDefinitionResourceProviderResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "function_id",
		},
		"resource type": {
			model: ResourceTypeModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				ResourceTypeIdentityModel: ResourceTypeIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
					ResourceTypeID:  types.StringValue("type"),
				},
				ResourceProviderID: types.StringUnknown(),
				Name:               types.StringUnknown(),
				DefaultFieldMapping: &ResourceTypeFieldMapping{
					Title:       types.StringValue("title"),
					Subtitle:    types.StringNull(),
					Description: types.StringNull(),
					ExternalURL: types.StringNull(),
				},
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: ResourceTypeResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := appDefinitionResourceTypeResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "name",
		},
		"role": {
			model: RoleModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				RoleIdentityModel: RoleIdentityModel{
					SpaceID: types.StringValue("space"),
					RoleID:  types.StringUnknown(),
				},
				Name:        types.StringValue("Role"),
				Description: types.StringNull(),
				Permissions: NewTypedMap(map[string]TypedList[types.String]{
					"ContentModel": NewTypedList([]types.String{types.StringValue("all"), types.StringValue("read")}),
				}),
				Policies: NewTypedList([]TypedObject[RolePolicyValue]{}),
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: RoleResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := roleResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: `permissions["ContentModel"]`,
		},
		"tag": {
			model: TagModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				TagIdentityModel: TagIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					TagID:         types.StringValue("tag"),
				},
				Name:       types.StringUnknown(),
				Visibility: types.StringValue("private"),
				Timeouts:   TimeoutsNull(),
			},
			resourceSchema: TagResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := tagResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "name",
		},
		"team space membership": {
			model: TeamSpaceMembershipModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
				TeamSpaceMembershipIdentityModel: TeamSpaceMembershipIdentityModel{
					SpaceID:               types.StringValue("space"),
					TeamSpaceMembershipID: types.StringUnknown(),
				},
				TeamID:   types.StringValue("team"),
				Admin:    types.BoolUnknown(),
				Roles:    []types.String{},
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: TeamSpaceMembershipResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := teamSpaceMembershipResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			expectedPath: "admin",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, requestCount := mutationRequestCountingClient(t)

			plan := tfsdk.Plan{Schema: test.resourceSchema}
			planDiags := plan.Set(ctx, test.model)
			require.False(t, planDiags.HasError(), planDiags.Errors())

			request := resource.CreateRequest{Plan: plan}

			if test.config != nil {
				configPlan := tfsdk.Plan{Schema: test.resourceSchema}
				configDiags := configPlan.Set(ctx, test.config)
				require.False(t, configDiags.HasError(), configDiags.Errors())

				request.Config = tfsdk.Config{Raw: configPlan.Raw, Schema: test.resourceSchema}
			}

			response := resource.CreateResponse{State: tfsdk.State{Schema: test.resourceSchema}}
			test.create(client, request, &response)

			require.True(t, response.Diagnostics.HasError())
			assert.Contains(t, mutationDiagnosticPaths(t, response.Diagnostics), test.expectedPath)
			assert.Zero(t, requestCount.Load())
		})
	}
}

func TestUpdateRequestConversionErrorsStopBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	for name, test := range map[string]struct {
		model          any
		config         any
		resourceSchema schema.Schema
		update         func(*cm.Client, resource.UpdateRequest, *resource.UpdateResponse)
		expectedPath   string
	}{
		"app signing secret": {
			model: AppSigningSecretModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("organization/app")},
				AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
				},
				Value:    types.StringUnknown(),
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: AppSigningSecretResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "value",
		},
		"delivery API key": {
			model: DeliveryAPIKeyModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/key")},
				DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
					SpaceID:  types.StringValue("space"),
					APIKeyID: types.StringValue("key"),
				},
				Name:            types.StringUnknown(),
				Description:     types.StringNull(),
				Environments:    NewTypedList([]types.String{}),
				AccessToken:     types.StringValue("redacted"),
				PreviewAPIKeyID: types.StringValue("preview"),
				Timeouts:        TimeoutsNull(),
			},
			config: DeliveryAPIKeyModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringNull()},
				DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
					SpaceID:  types.StringValue("space"),
					APIKeyID: types.StringNull(),
				},
				Name:            types.StringValue("configured key"),
				Description:     types.StringNull(),
				Environments:    NewTypedList([]types.String{}),
				AccessToken:     types.StringNull(),
				PreviewAPIKeyID: types.StringNull(),
				Timeouts:        TimeoutsNull(),
			},
			resourceSchema: DeliveryAPIKeyResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := deliveryAPIKeyResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "name",
		},
		"environment": {
			model: EnvironmentModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/environment")},
				EnvironmentIdentityModel: EnvironmentIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
				},
				Name:                types.StringUnknown(),
				Status:              types.StringValue("ready"),
				SourceEnvironmentID: types.StringNull(),
				Timeouts:            TimeoutsNull(),
			},
			resourceSchema: EnvironmentResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := environmentResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "name",
		},
		"environment alias": {
			model: EnvironmentAliasModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/alias")},
				EnvironmentAliasIdentityModel: EnvironmentAliasIdentityModel{
					SpaceID:            types.StringValue("space"),
					EnvironmentAliasID: types.StringValue("alias"),
				},
				TargetEnvironmentID: types.StringUnknown(),
				Timeouts:            TimeoutsNull(),
			},
			resourceSchema: EnvironmentAliasResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := environmentAliasResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "target_environment_id",
		},
		"resource provider": {
			model: ResourceProviderModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("organization/app")},
				ResourceProviderIdentityModel: ResourceProviderIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
				},
				ResourceProviderID: types.StringValue("provider"),
				FunctionID:         types.StringUnknown(),
				Timeouts:           TimeoutsNull(),
			},
			resourceSchema: ResourceProviderResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := appDefinitionResourceProviderResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "function_id",
		},
		"resource type": {
			model: ResourceTypeModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("organization/app/type")},
				ResourceTypeIdentityModel: ResourceTypeIdentityModel{
					OrganizationID:  types.StringValue("organization"),
					AppDefinitionID: types.StringValue("app"),
					ResourceTypeID:  types.StringValue("type"),
				},
				ResourceProviderID: types.StringValue("provider"),
				Name:               types.StringUnknown(),
				DefaultFieldMapping: &ResourceTypeFieldMapping{
					Title:       types.StringValue("title"),
					Subtitle:    types.StringNull(),
					Description: types.StringNull(),
					ExternalURL: types.StringNull(),
				},
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: ResourceTypeResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := appDefinitionResourceTypeResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "name",
		},
		"tag": {
			model: TagModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/environment/tag")},
				TagIdentityModel: TagIdentityModel{
					SpaceID:       types.StringValue("space"),
					EnvironmentID: types.StringValue("environment"),
					TagID:         types.StringValue("tag"),
				},
				Name:       types.StringUnknown(),
				Visibility: types.StringValue("private"),
				Timeouts:   TimeoutsNull(),
			},
			resourceSchema: TagResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := tagResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "name",
		},
		"team space membership": {
			model: TeamSpaceMembershipModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/membership")},
				TeamSpaceMembershipIdentityModel: TeamSpaceMembershipIdentityModel{
					SpaceID:               types.StringValue("space"),
					TeamSpaceMembershipID: types.StringValue("membership"),
				},
				TeamID:   types.StringValue("team"),
				Admin:    types.BoolUnknown(),
				Roles:    []types.String{},
				Timeouts: TimeoutsNull(),
			},
			resourceSchema: TeamSpaceMembershipResourceSchema(ctx),
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := teamSpaceMembershipResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "admin",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, requestCount := mutationRequestCountingClient(t)

			plan := tfsdk.Plan{Schema: test.resourceSchema}
			planDiags := plan.Set(ctx, test.model)
			require.False(t, planDiags.HasError(), planDiags.Errors())

			request := resource.UpdateRequest{
				Plan:  plan,
				State: tfsdk.State{Raw: plan.Raw, Schema: test.resourceSchema},
			}

			if test.config != nil {
				configPlan := tfsdk.Plan{Schema: test.resourceSchema}
				configDiags := configPlan.Set(ctx, test.config)
				require.False(t, configDiags.HasError(), configDiags.Errors())

				request.Config = tfsdk.Config{Raw: configPlan.Raw, Schema: test.resourceSchema}
			}

			response := resource.UpdateResponse{State: tfsdk.State{Schema: test.resourceSchema}}
			test.update(client, request, &response)

			require.True(t, response.Diagnostics.HasError())
			assert.Contains(t, mutationDiagnosticPaths(t, response.Diagnostics), test.expectedPath)
			assert.Zero(t, requestCount.Load())
		})
	}
}

func TestEntryUpdateRequestConversionErrorStopsBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	client, requestCount := mutationRequestCountingClient(t)
	implementation := entryResource{providerData: ContentfulProviderData{client: client}}
	model := EntryModel{
		EntryIdentityModel: EntryIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			EntryID:       types.StringValue("entry"),
		},
		ContentTypeID: types.StringValue("content-type"),
		Fields:        NewTypedMapUnknown[jsontypes.Normalized](),
		Metadata: NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{}),
			Tags:     NewTypedList([]types.String{}),
		}),
	}
	diags := diag.Diagnostics{}

	implementation.updateEntry(t.Context(), model, 1, &diags)

	require.True(t, diags.HasError())
	assert.Contains(t, mutationDiagnosticPaths(t, diags), "fields")
	assert.Zero(t, requestCount.Load())
}

//nolint:maintidx // One table intentionally verifies both lifecycles for every configuration-aware converter.
func TestUnknownPlannedConfigurationOwnedValueStopsBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	deliveryAPIKeyPlan := DeliveryAPIKeyModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
			SpaceID:  types.StringValue("space"),
			APIKeyID: types.StringUnknown(),
		},
		Name:            types.StringValue("Key"),
		Description:     types.StringNull(),
		Environments:    NewTypedListUnknown[types.String](),
		AccessToken:     types.StringUnknown(),
		PreviewAPIKeyID: types.StringUnknown(),
		Timeouts:        TimeoutsNull(),
	}
	deliveryAPIKeyConfig := deliveryAPIKeyPlan
	deliveryAPIKeyConfig.ID = types.StringNull()
	deliveryAPIKeyConfig.APIKeyID = types.StringNull()
	deliveryAPIKeyConfig.Environments = NewTypedList([]types.String{})
	deliveryAPIKeyConfig.AccessToken = types.StringNull()
	deliveryAPIKeyConfig.PreviewAPIKeyID = types.StringNull()

	appDefinitionPlan := AppDefinitionResourceModel{
		AppDefinitionBaseModel: AppDefinitionBaseModel{
			IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
			AppDefinitionIdentityModel: AppDefinitionIdentityModel{
				OrganizationID:  types.StringValue("organization"),
				AppDefinitionID: types.StringUnknown(),
			},
			Name:      types.StringValue("App"),
			Src:       types.StringUnknown(),
			BundleID:  types.StringNull(),
			Locations: []AppDefinitionLocationsItem{},
		},
		Timeouts: TimeoutsNull(),
	}
	appDefinitionConfig := appDefinitionPlan
	appDefinitionConfig.ID = types.StringNull()
	appDefinitionConfig.AppDefinitionID = types.StringNull()
	appDefinitionConfig.Src = types.StringValue("https://example.com/app.js")

	contentTypePlan := ContentTypeModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		ContentTypeIdentityModel: ContentTypeIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			ContentTypeID: types.StringValue("content-type"),
		},
		Name:         types.StringValue("Content type"),
		Description:  types.StringValue("Description"),
		DisplayField: types.StringValue("title"),
		Fields:       NewTypedList([]TypedObject[ContentTypeFieldValue]{}),
		Metadata:     NewTypedObjectUnknown[ContentTypeMetadataValue](),
		Timeouts:     TimeoutsNull(),
	}
	contentTypeConfig := contentTypePlan
	contentTypeConfig.ID = types.StringNull()
	contentTypeConfig.Metadata = NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedNull(),
		Taxonomy:    NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{}),
	})

	extensionPlan := ExtensionModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		ExtensionIdentityModel: ExtensionIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			ExtensionID:   types.StringValue("extension"),
		},
		Extension: &ExtensionConfiguration{
			Name:       types.StringValue("Extension"),
			Src:        types.StringNull(),
			SrcDoc:     types.StringUnknown(),
			FieldTypes: []AppDefinitionLocationFieldTypesItem{},
			Sidebar:    types.BoolValue(false),
		},
		Parameters: jsontypes.NewNormalizedNull(),
		Timeouts:   TimeoutsNull(),
	}
	extensionConfig := extensionPlan
	extensionConfig.ID = types.StringNull()
	extensionConfig.Extension = &ExtensionConfiguration{
		Name:       types.StringValue("Extension"),
		Src:        types.StringNull(),
		SrcDoc:     types.StringValue("<!doctype html>"),
		FieldTypes: []AppDefinitionLocationFieldTypesItem{},
		Sidebar:    types.BoolValue(false),
	}

	webhookPlan := WebhookModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		WebhookIdentityModel: WebhookIdentityModel{
			SpaceID:   types.StringValue("space"),
			WebhookID: types.StringUnknown(),
		},
		Name:              types.StringValue("Webhook"),
		URL:               types.StringValue("https://example.com/webhook"),
		Topics:            NewTypedListNull[types.String](),
		Filters:           NewTypedListNull[TypedObject[WebhookFilterValue]](),
		HTTPBasicPassword: types.StringNull(),
		HTTPBasicUsername: types.StringNull(),
		Headers:           NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
		Transformation:    NewTypedObjectNull[WebhookTransformationValue](),
		Active:            types.BoolValue(true),
		Timeouts:          TimeoutsNull(),
	}
	webhookConfig := webhookPlan
	webhookConfig.ID = types.StringNull()
	webhookConfig.WebhookID = types.StringNull()
	webhookConfig.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{})

	spaceEnablementsPlan := SpaceEnablementsModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		SpaceEnablementsIdentityModel: SpaceEnablementsIdentityModel{
			SpaceID: types.StringValue("space"),
		},
		CrossSpaceLinks:   types.BoolUnknown(),
		SpaceTemplates:    types.BoolNull(),
		StudioExperiences: types.BoolNull(),
		SuggestConcepts:   types.BoolNull(),
		Timeouts:          spaceEnablementsTimeoutsNull(),
	}
	spaceEnablementsConfig := spaceEnablementsPlan
	spaceEnablementsConfig.ID = types.StringNull()
	spaceEnablementsConfig.CrossSpaceLinks = types.BoolValue(false)

	for name, test := range map[string]struct {
		plan           any
		config         any
		resourceSchema schema.Schema
		create         func(*cm.Client, resource.CreateRequest, *resource.CreateResponse)
		update         func(*cm.Client, resource.UpdateRequest, *resource.UpdateResponse)
		expectedPath   string
	}{
		"delivery API key": {
			plan:           deliveryAPIKeyPlan,
			config:         deliveryAPIKeyConfig,
			resourceSchema: DeliveryAPIKeyResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := deliveryAPIKeyResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := deliveryAPIKeyResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "environments",
		},
		"app definition": {
			plan:           appDefinitionPlan,
			config:         appDefinitionConfig,
			resourceSchema: AppDefinitionResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := appDefinitionResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := appDefinitionResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "src",
		},
		"content type": {
			plan:           contentTypePlan,
			config:         contentTypeConfig,
			resourceSchema: ContentTypeResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := contentTypeResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := contentTypeResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "metadata",
		},
		"extension": {
			plan:           extensionPlan,
			config:         extensionConfig,
			resourceSchema: ExtensionResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := extensionResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := extensionResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "extension.srcdoc",
		},
		"webhook": {
			plan:           webhookPlan,
			config:         webhookConfig,
			resourceSchema: WebhookResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := webhookResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := webhookResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "headers",
		},
		"space enablements": {
			plan:           spaceEnablementsPlan,
			config:         spaceEnablementsConfig,
			resourceSchema: SpaceEnablementsResourceSchema(ctx),
			create: func(client *cm.Client, request resource.CreateRequest, response *resource.CreateResponse) {
				implementation := spaceEnablementsResource{providerData: ContentfulProviderData{client: client}}
				implementation.Create(ctx, request, response)
			},
			update: func(client *cm.Client, request resource.UpdateRequest, response *resource.UpdateResponse) {
				implementation := spaceEnablementsResource{providerData: ContentfulProviderData{client: client}}
				implementation.Update(ctx, request, response)
			},
			expectedPath: "cross_space_links",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := tfsdk.Plan{Schema: test.resourceSchema}
			planDiags := plan.Set(ctx, test.plan)
			require.False(t, planDiags.HasError(), planDiags.Errors())

			configPlan := tfsdk.Plan{Schema: test.resourceSchema}
			configDiags := configPlan.Set(ctx, test.config)
			require.False(t, configDiags.HasError(), configDiags.Errors())

			config := tfsdk.Config{Raw: configPlan.Raw, Schema: test.resourceSchema}

			client, requestCount := mutationRequestCountingClient(t)
			createResponse := resource.CreateResponse{State: tfsdk.State{Schema: test.resourceSchema}}
			test.create(client, resource.CreateRequest{Config: config, Plan: plan}, &createResponse)

			require.True(t, createResponse.Diagnostics.HasError())
			assert.Contains(t, mutationDiagnosticPaths(t, createResponse.Diagnostics), test.expectedPath)
			assert.Zero(t, requestCount.Load())

			updateResponse := resource.UpdateResponse{State: tfsdk.State{Schema: test.resourceSchema}}
			test.update(client, resource.UpdateRequest{Config: config, Plan: plan}, &updateResponse)

			require.True(t, updateResponse.Diagnostics.HasError())
			assert.Contains(t, mutationDiagnosticPaths(t, updateResponse.Diagnostics), test.expectedPath)
			assert.Zero(t, requestCount.Load())
		})
	}
}

func mutationDiagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		if ok {
			paths = append(paths, withPath.Path().String())
		}
	}

	return paths
}

func personalAccessTokenTimeoutsNull() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"delete": types.StringType,
	})}
}

func spaceEnablementsTimeoutsNull() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"update": types.StringType,
	})}
}
