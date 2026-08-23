package provider_test

import (
	"context"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	frameworklist "github.com/hashicorp/terraform-plugin-framework/list"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeResourceLegacyStateWithoutRefreshConservativelyPlansActivation(t *testing.T) {
	t.Parallel()
	testAccContentTypeResourceLegacyStateUpgrade(t, true)
}

func TestAccContentTypeResourceLegacyStateNormalRefreshObservesPublicationState(t *testing.T) {
	t.Parallel()
	testAccContentTypeResourceLegacyStateUpgrade(t, false)
}

func testAccContentTypeResourceLegacyStateUpgrade(t *testing.T, noRefresh bool) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	contentTypeID := "legacy-refresh"
	if noRefresh {
		contentTypeID = "legacy-no-refresh"
	}

	contentType := seedActivatedLegacyContentType(t, server, contentTypeID)
	handler := &contentTypeActivationTestHandler{delegate: server}
	testserver := httptest.NewServer(handler)
	t.Cleanup(testserver.Close)
	options := ContentfulProviderOptionsWithHTTPTestServer(testserver)
	variables := config.Variables{
		"space_id":             config.StringVariable("space"),
		"environment_id":       config.StringVariable("environment"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	var additionalCLIOptions *testingresource.AdditionalCLIOptions
	if noRefresh {
		additionalCLIOptions = &testingresource.AdditionalCLIOptions{
			Plan: testingresource.PlanOptions{NoRefresh: true},
		}
	}

	refreshedPlanChecks := []plancheck.PlanCheck{
		plancheck.ExpectKnownValue(
			"contentful_content_type.test",
			tfjsonpath.New("published_version"),
			knownvalue.Int64Exact(int64(contentType.Sys.PublishedVersion.Value)),
		),
		plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
	}
	nonRefreshPlanChecks := []plancheck.PlanCheck{
		plancheck.ExpectKnownValue(
			"contentful_content_type.test",
			tfjsonpath.New("published_version"),
			knownvalue.Int64Exact(int64(contentType.Sys.Version)),
		),
		plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
	}

	var currentProviderStep testingresource.TestStep
	if noRefresh {
		currentProviderStep = testingresource.TestStep{
			ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
			ConfigDirectory:          config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables:          variables,
			ConfigPlanChecks: testingresource.ConfigPlanChecks{
				PreApply: nonRefreshPlanChecks,
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"contentful_content_type.test",
					tfjsonpath.New("published_version"),
					knownvalue.Int64Exact(int64(contentType.Sys.Version)),
				),
			},
			Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{int64(contentType.Sys.Version)}),
		}
	} else {
		currentProviderStep = testingresource.TestStep{
			ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
			ConfigDirectory:          config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables:          variables,
			ConfigPlanChecks: testingresource.ConfigPlanChecks{
				PreApply: refreshedPlanChecks,
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"contentful_content_type.test",
					tfjsonpath.New("published_version"),
					knownvalue.Int64Exact(int64(contentType.Sys.PublishedVersion.Value)),
				),
			},
		}
	}

	steps := []testingresource.TestStep{
		{
			ProtoV6ProviderFactories: legacyContentTypeProviderFactories(contentType, options...),
			ConfigDirectory:          config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables:          variables,
		},
		currentProviderStep,
	}
	if noRefresh {
		steps = append(steps, testingresource.TestStep{
			PreConfig: func() {
				additionalCLIOptions.Plan.NoRefresh = false

				handler.resetRequestHistory()
			},
			ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
			ConfigDirectory:          config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables:          variables,
			ConfigPlanChecks: testingresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectKnownValue(
					"contentful_content_type.test",
					tfjsonpath.New("published_version"),
					knownvalue.Int64Exact(int64(contentType.Sys.Version)),
				),
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		})
	}

	testingresource.Test(t, testingresource.TestCase{
		AdditionalCLIOptions: additionalCLIOptions,
		Steps:                steps,
	})
}

func seedActivatedLegacyContentType(t *testing.T, server *cmt.Server, contentTypeID string) cm.ContentType {
	t.Helper()

	request := cm.ContentTypeRequestData{
		Name:         "Test",
		Description:  cm.NewOptNilString("Test content type (" + contentTypeID + ")"),
		DisplayField: "name",
		Fields: []cm.ContentTypeRequestDataFieldsItem{
			{
				ID: "name", Name: "Name", Type: "Symbol",
				Required: cm.NewOptBool(true), Localized: cm.NewOptBool(false),
				Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false),
			},
			{
				ID: "flags", Name: "Flags", Type: "Array",
				Items:    cm.NewOptContentTypeRequestDataFieldsItemItems(cm.ContentTypeRequestDataFieldsItemItems{Type: cm.NewOptString("Symbol")}),
				Required: cm.NewOptBool(false), Localized: cm.NewOptBool(false),
				Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false),
			},
		},
	}
	putResponse, err := server.Handler().PutContentType(t.Context(), &request, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: 1,
	})
	require.NoError(t, err)

	put, ok := putResponse.(*cm.ContentTypeStatusCode)
	require.True(t, ok)

	activationResponse, err := server.Handler().ActivateContentType(t.Context(), cm.ActivateContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: put.Response.Sys.Version,
	})
	require.NoError(t, err)

	activation, ok := activationResponse.(*cm.ContentTypeStatusCode)
	require.True(t, ok)

	return activation.Response
}

type legacyContentTypeProvider struct {
	*ContentfulProvider

	contentType cm.ContentType
}

func (p *legacyContentTypeProvider) ListResources(context.Context) []func() frameworklist.ListResource {
	return nil
}

func (p *legacyContentTypeProvider) Resources(context.Context) []func() frameworkresource.Resource {
	return []func() frameworkresource.Resource{
		func() frameworkresource.Resource { return &legacyContentTypeResource{contentType: p.contentType} },
	}
}

func legacyContentTypeProviderFactories(contentType cm.ContentType, options ...Option) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"contentful": providerserver.NewProtocol6WithError(func() frameworkprovider.Provider {
			return &legacyContentTypeProvider{ContentfulProvider: New("test", options...), contentType: contentType}
		}()),
	}
}

type legacyContentTypeResource struct {
	contentType cm.ContentType
}

func (r *legacyContentTypeResource) Metadata(_ context.Context, req frameworkresource.MetadataRequest, resp *frameworkresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_type"
}

func (r *legacyContentTypeResource) Schema(ctx context.Context, _ frameworkresource.SchemaRequest, resp *frameworkresource.SchemaResponse) {
	resp.Schema = ContentTypeResourceSchema(ctx)
	resp.Schema.Version = 0
	delete(resp.Schema.Attributes, "published_version")
}

func (r *legacyContentTypeResource) Create(ctx context.Context, req frameworkresource.CreateRequest, resp *frameworkresource.CreateResponse) {
	var plan legacyContentTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	current, projectionDiags := NewContentTypeResourceModelFromResponse(ctx, r.contentType)
	resp.Diagnostics.Append(projectionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := legacyContentTypeModel{
		IDIdentityModel:          current.IDIdentityModel,
		ContentTypeIdentityModel: current.ContentTypeIdentityModel,
		Name:                     current.Name,
		Description:              current.Description,
		DisplayField:             current.DisplayField,
		Fields:                   current.Fields,
		Metadata:                 current.Metadata,
		Timeouts:                 plan.Timeouts,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", r.contentType.Sys.Version)...)
}

func (r *legacyContentTypeResource) Read(context.Context, frameworkresource.ReadRequest, *frameworkresource.ReadResponse) {
}

func (r *legacyContentTypeResource) Update(context.Context, frameworkresource.UpdateRequest, *frameworkresource.UpdateResponse) {
}

func (r *legacyContentTypeResource) Delete(context.Context, frameworkresource.DeleteRequest, *frameworkresource.DeleteResponse) {
}

type legacyContentTypeModel struct {
	IDIdentityModel
	ContentTypeIdentityModel

	Name         types.String                                  `tfsdk:"name"`
	Description  types.String                                  `tfsdk:"description"`
	DisplayField types.String                                  `tfsdk:"display_field"`
	Fields       TypedList[TypedObject[ContentTypeFieldValue]] `tfsdk:"fields"`
	Metadata     TypedObject[ContentTypeMetadataValue]         `tfsdk:"metadata"`
	Timeouts     timeouts.Value                                `tfsdk:"timeouts"`
}
