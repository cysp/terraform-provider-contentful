//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionResourceProviderResource)(nil)
)

//nolint:ireturn
func NewResourceProviderResource() resource.Resource {
	return &appDefinitionResourceProviderResource{}
}

type appDefinitionResourceProviderResource struct {
	providerData ContentfulProviderData
}

func resourceProviderIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id"}
}

func (r *appDefinitionResourceProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_provider"
}

func (r *appDefinitionResourceProviderResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceProviderResourceSchema(ctx)
}

func (r *appDefinitionResourceProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionResourceProviderResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(resourceProviderIdentityAttributeNames())
}

func (r *appDefinitionResourceProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, resourceProviderIdentityAttributeNames(), req, resp)
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.PutResourceProviderParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	request, requestDiags := plan.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "resource_provider.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data ResourceProviderModel

	switch response := response.(type) {
	case *cm.ResourceProviderStatusCode:
		responseModel, responseModelDiags := NewResourceProviderResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceProviderIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.GetResourceProviderParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetResourceProvider(ctx, params)

	tflog.Info(ctx, "resource_provider.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	var data ResourceProviderModel

	switch response := response.(type) {
	case *cm.ResourceProvider:
		responseModel, responseModelDiags := NewResourceProviderResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceProviderIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.PutResourceProviderParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	request, requestDiags := plan.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "resource_provider.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data ResourceProviderModel

	switch response := response.(type) {
	case *cm.ResourceProviderStatusCode:
		responseModel, responseModelDiags := NewResourceProviderResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceProviderIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.DeleteResourceProvider(ctx, cm.DeleteResourceProviderParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Resource provider definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			handled = true
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
