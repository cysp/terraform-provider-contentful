//nolint:testpackage // Lifecycle methods and their package-local implementation are exercised directly.
package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errAppSigningSecretTestTransport = errors.New("transport exposed sensitive value")

func TestAppSigningSecretLifecycleRuntimeOutputExcludesValues(t *testing.T) {
	t.Parallel()

	initialValue := appSigningSecretLogSentinel("CREATE")
	updatedValue := appSigningSecretLogSentinel("UPDATE")

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	var logOutput bytes.Buffer

	ctx := tflogtest.RootLogger(t.Context(), &logOutput)
	resourceSchema := AppSigningSecretResourceSchema(ctx)
	implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}

	createPlan := appSigningSecretTestPlan(ctx, t, resourceSchema, appSigningSecretTestModel(initialValue))
	createResponse := resource.CreateResponse{
		State:    tfsdk.State{Schema: resourceSchema},
		Identity: appSigningSecretTestIdentity(ctx),
	}
	implementation.Create(ctx, resource.CreateRequest{
		Config: tfsdk.Config(createPlan),
		Plan:   createPlan,
	}, &createResponse)
	require.False(t, createResponse.Diagnostics.HasError(), createResponse.Diagnostics.Errors())

	readResponse := resource.ReadResponse{
		State:    createResponse.State,
		Identity: createResponse.Identity,
	}
	implementation.Read(ctx, resource.ReadRequest{
		State:    createResponse.State,
		Identity: createResponse.Identity,
	}, &readResponse)
	require.False(t, readResponse.Diagnostics.HasError(), readResponse.Diagnostics.Errors())

	var updateModel AppSigningSecretModel
	require.False(t, readResponse.State.Get(ctx, &updateModel).HasError())
	updateModel.Value = types.StringValue(updatedValue)
	updatePlan := appSigningSecretTestPlan(ctx, t, resourceSchema, updateModel)
	updateResponse := resource.UpdateResponse{
		State:    tfsdk.State{Raw: updatePlan.Raw, Schema: resourceSchema},
		Identity: readResponse.Identity,
	}
	implementation.Update(ctx, resource.UpdateRequest{
		Config:   tfsdk.Config(updatePlan),
		Plan:     updatePlan,
		State:    readResponse.State,
		Identity: readResponse.Identity,
	}, &updateResponse)
	require.False(t, updateResponse.Diagnostics.HasError(), updateResponse.Diagnostics.Errors())

	deleteResponse := resource.DeleteResponse{
		State:    updateResponse.State,
		Identity: updateResponse.Identity,
	}
	implementation.Delete(ctx, resource.DeleteRequest{
		State:    updateResponse.State,
		Identity: updateResponse.Identity,
	}, &deleteResponse)
	require.Empty(t, deleteResponse.Diagnostics)

	deleted, err := client.GetAppSigningSecret(ctx, cm.GetAppSigningSecretParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
	})
	require.NoError(t, err)

	deletedStatus, ok := deleted.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, deletedStatus.GetStatusCode())

	assertAppSigningSecretOperationLogs(t, logOutput.Bytes(),
		"app_signing_secret.create",
		"app_signing_secret.read",
		"app_signing_secret.update",
	)
	assertAppSigningSecretRuntimeOutputExcludes(t, logOutput.String(), []diag.Diagnostics{
		createResponse.Diagnostics,
		readResponse.Diagnostics,
		updateResponse.Diagnostics,
		deleteResponse.Diagnostics,
	}, initialValue, updatedValue)
}

func TestAppSigningSecretLifecycleErrorOutputRedactsValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		message string
		run     func(*testing.T, *appSigningSecretResource, tfsdk.Plan, tfsdk.State, *tfsdk.ResourceIdentity, *bytes.Buffer) diag.Diagnostics
	}{
		"create": {
			message: "app_signing_secret.create",
			run: func(t *testing.T, implementation *appSigningSecretResource, plan tfsdk.Plan, _ tfsdk.State, identity *tfsdk.ResourceIdentity, logs *bytes.Buffer) diag.Diagnostics {
				t.Helper()

				response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: identity}
				implementation.Create(tflogtest.RootLogger(t.Context(), logs), resource.CreateRequest{
					Config: tfsdk.Config(plan),
					Plan:   plan,
				}, &response)

				return response.Diagnostics
			},
		},
		"read": {
			message: "app_signing_secret.read",
			run: func(t *testing.T, implementation *appSigningSecretResource, _ tfsdk.Plan, state tfsdk.State, identity *tfsdk.ResourceIdentity, logs *bytes.Buffer) diag.Diagnostics {
				t.Helper()

				response := resource.ReadResponse{State: state, Identity: identity}
				implementation.Read(tflogtest.RootLogger(t.Context(), logs), resource.ReadRequest{State: state, Identity: identity}, &response)

				return response.Diagnostics
			},
		},
		"update": {
			message: "app_signing_secret.update",
			run: func(t *testing.T, implementation *appSigningSecretResource, plan tfsdk.Plan, state tfsdk.State, identity *tfsdk.ResourceIdentity, logs *bytes.Buffer) diag.Diagnostics {
				t.Helper()

				response := resource.UpdateResponse{State: tfsdk.State(plan), Identity: identity}
				implementation.Update(tflogtest.RootLogger(t.Context(), logs), resource.UpdateRequest{
					Config:   tfsdk.Config(plan),
					Plan:     plan,
					State:    state,
					Identity: identity,
				}, &response)

				return response.Diagnostics
			},
		},
		"delete": {
			run: func(t *testing.T, implementation *appSigningSecretResource, _ tfsdk.Plan, state tfsdk.State, identity *tfsdk.ResourceIdentity, logs *bytes.Buffer) diag.Diagnostics {
				t.Helper()

				response := resource.DeleteResponse{State: state, Identity: identity}
				implementation.Delete(tflogtest.RootLogger(t.Context(), logs), resource.DeleteRequest{State: state, Identity: identity}, &response)

				return response.Diagnostics
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value := appSigningSecretLogSentinel(strings.ToUpper(name))
			client, err := cm.NewClient(
				"https://contentful.invalid",
				cm.NewAccessTokenSecuritySource("access-token"),
				cm.WithClient(&http.Client{Transport: appSigningSecretRoundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("%w: %s", errAppSigningSecretTestTransport, value)
				})}),
			)
			require.NoError(t, err)

			ctx := t.Context()
			resourceSchema := AppSigningSecretResourceSchema(ctx)
			model := appSigningSecretTestModel(value)
			plan := appSigningSecretTestPlan(ctx, t, resourceSchema, model)
			state := tfsdk.State{Schema: resourceSchema}
			require.False(t, state.Set(ctx, &model).HasError())
			identity := appSigningSecretTestIdentity(ctx)
			require.False(t, identity.Set(ctx, &model.AppSigningSecretIdentityModel).HasError())

			var logOutput bytes.Buffer

			implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
			diagnostics := test.run(t, &implementation, plan, state, identity, &logOutput)

			require.True(t, diagnostics.HasError())

			if test.message != "" {
				assertAppSigningSecretOperationLogs(t, logOutput.Bytes(), test.message)
				assert.Contains(t, logOutput.String(), "***", "the actual logged error must be masked")
			}

			assertAppSigningSecretRuntimeOutputExcludes(t, logOutput.String(), []diag.Diagnostics{diagnostics}, value)
		})
	}
}

func TestAppSigningSecretCMAErrorDiagnosticRedactsValue(t *testing.T) {
	t.Parallel()

	value := appSigningSecretLogSentinel("CMA_ERROR")
	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		writeErr := writeAppSigningSecretTestCMAErrorResponse(
			responseWriter,
			http.StatusUnprocessableEntity,
			"ValidationFailed",
			"CMA rejected "+value,
		)
		if writeErr != nil {
			t.Errorf("write Contentful error response: %s", writeErr)
		}
	}))
	t.Cleanup(testServer.Close)

	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	var logOutput bytes.Buffer

	ctx := tflogtest.RootLogger(t.Context(), &logOutput)
	resourceSchema := AppSigningSecretResourceSchema(ctx)
	plan := appSigningSecretTestPlan(ctx, t, resourceSchema, appSigningSecretTestModel(value))
	response := resource.CreateResponse{
		State:    tfsdk.State{Schema: resourceSchema},
		Identity: appSigningSecretTestIdentity(ctx),
	}
	implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
	implementation.Create(ctx, resource.CreateRequest{
		Config: tfsdk.Config(plan),
		Plan:   plan,
	}, &response)

	require.True(t, response.Diagnostics.HasError())
	assertAppSigningSecretOperationLogs(t, logOutput.Bytes(), "app_signing_secret.create")
	assertAppSigningSecretRuntimeOutputExcludes(t, logOutput.String(), []diag.Diagnostics{response.Diagnostics}, value)
}

func TestAppSigningSecretDeleteCMAErrorDiagnosticsRedactValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		statusCode int
		errorID    string
		hasError   bool
	}{
		"not found warning": {
			statusCode: http.StatusNotFound,
			errorID:    cm.ErrorSysIDNotFound,
		},
		"CMA error": {
			statusCode: http.StatusUnprocessableEntity,
			errorID:    "ValidationFailed",
			hasError:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value := appSigningSecretLogSentinel("DELETE_CMA_ERROR")
			testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
				writeErr := writeAppSigningSecretTestCMAErrorResponse(
					responseWriter,
					test.statusCode,
					test.errorID,
					"CMA exposed "+value,
				)
				if writeErr != nil {
					t.Errorf("write Contentful error response: %s", writeErr)
				}
			}))
			t.Cleanup(testServer.Close)

			client, err := cm.NewClient(
				testServer.URL,
				cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
				cm.WithClient(testServer.Client()),
			)
			require.NoError(t, err)

			ctx := t.Context()
			resourceSchema := AppSigningSecretResourceSchema(ctx)
			model := appSigningSecretTestModel(value)
			state := tfsdk.State{Schema: resourceSchema}
			require.False(t, state.Set(ctx, &model).HasError())

			var logOutput bytes.Buffer

			implementation := appSigningSecretResource{providerData: ContentfulProviderData{client: client}}
			response := resource.DeleteResponse{State: state, Identity: appSigningSecretTestIdentity(ctx)}
			implementation.Delete(tflogtest.RootLogger(ctx, &logOutput), resource.DeleteRequest{State: state}, &response)

			assert.Equal(t, test.hasError, response.Diagnostics.HasError(), response.Diagnostics)
			require.Len(t, response.Diagnostics, 1)
			assertAppSigningSecretRuntimeOutputExcludes(t, logOutput.String(), []diag.Diagnostics{response.Diagnostics}, value)
		})
	}
}

type appSigningSecretRoundTripFunc func(*http.Request) (*http.Response, error)

func (f appSigningSecretRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func writeAppSigningSecretTestCMAErrorResponse(responseWriter http.ResponseWriter, statusCode int, id, message string) error {
	response := cmt.NewContentfulManagementError(id, &message, nil)

	body, err := response.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal Contentful error response: %w", err)
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	_, err = responseWriter.Write(body)
	if err != nil {
		return fmt.Errorf("write Contentful error response: %w", err)
	}

	return nil
}

func appSigningSecretLogSentinel(label string) string {
	prefix := "DO_NOT_LOG_" + label + "_APP_SIGNING_SECRET_"

	return prefix + strings.Repeat("X", 64-len(prefix))
}

func appSigningSecretTestModel(value string) AppSigningSecretModel {
	return AppSigningSecretModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringValue("organization/app-definition")},
		AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
			OrganizationID:  types.StringValue("organization"),
			AppDefinitionID: types.StringValue("app-definition"),
		},
		Value:    types.StringValue(value),
		Timeouts: TimeoutsNull(),
	}
}

func appSigningSecretTestPlan(ctx context.Context, t *testing.T, resourceSchema schema.Schema, model AppSigningSecretModel) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, &model).HasError())

	return plan
}

func appSigningSecretTestIdentity(ctx context.Context) *tfsdk.ResourceIdentity {
	identitySchema := identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
		"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
	}}

	return &tfsdk.ResourceIdentity{
		Schema: identitySchema,
		Raw:    tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil),
	}
}

func assertAppSigningSecretOperationLogs(t *testing.T, output []byte, expectedMessages ...string) {
	t.Helper()

	entries, err := tflogtest.MultilineJSONDecode(bytes.NewReader(output))
	require.NoError(t, err)

	messages := make([]string, 0, len(entries))
	for _, entry := range entries {
		message, ok := entry["@message"].(string)
		if ok {
			messages = append(messages, message)
		}
	}

	for _, expectedMessage := range expectedMessages {
		assert.Contains(t, messages, expectedMessage)
	}
}

func assertAppSigningSecretRuntimeOutputExcludes(t *testing.T, logs string, diagnosticSets []diag.Diagnostics, values ...string) {
	t.Helper()

	var output strings.Builder
	output.WriteString(logs)

	for _, diagnostics := range diagnosticSets {
		for _, diagnostic := range diagnostics {
			output.WriteString("\n")
			output.WriteString(diagnostic.Summary())
			output.WriteString("\n")
			output.WriteString(diagnostic.Detail())
		}
	}

	for _, value := range values {
		assert.NotContains(t, output.String(), value)
	}
}
