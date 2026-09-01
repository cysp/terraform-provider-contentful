package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionResourceTypeResource)(nil)
)

//nolint:ireturn
func NewResourceTypeResource() resource.Resource {
	return &appDefinitionResourceTypeResource{}
}

type appDefinitionResourceTypeResource struct {
	providerData ContentfulProviderData
}

func resourceTypeIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id", "resource_type_id"}
}

func (r *appDefinitionResourceTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_type"
}

func (r *appDefinitionResourceTypeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceTypeResourceSchema(ctx)
}

func (r *appDefinitionResourceTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionResourceTypeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(resourceTypeIdentityAttributeNames())
}

func (r *appDefinitionResourceTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, resourceTypeIdentityAttributeNames(), req, resp)
}

//nolint:dupl
func (r *appDefinitionResourceTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceTypeModel

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

	params := cm.PutResourceTypeParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
		ResourceTypeID:  plan.ResourceTypeID.ValueString(),
	}

	request, requestDiags := plan.ToResourceTypeData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceType(ctx, &request, params)

	tflog.Info(ctx, "resource_type.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data ResourceTypeModel

	switch response := response.(type) {
	case *cm.ResourceTypeStatusCode:
		responseModel, responseModelDiags := NewResourceTypeResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceTypeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appDefinitionResourceTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceTypeModel

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

	params := cm.GetResourceTypeParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
		ResourceTypeID:  state.ResourceTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetResourceType(ctx, params)

	tflog.Info(ctx, "resource_type.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	var data ResourceTypeModel

	switch response := response.(type) {
	case *cm.ResourceType:
		responseModel, responseModelDiags := NewResourceTypeResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceTypeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appDefinitionResourceTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceTypeModel

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

	params := cm.PutResourceTypeParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
		ResourceTypeID:  plan.ResourceTypeID.ValueString(),
	}

	request, requestDiags := plan.ToResourceTypeData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceType(ctx, &request, params)

	tflog.Info(ctx, "resource_type.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data ResourceTypeModel

	switch response := response.(type) {
	case *cm.ResourceTypeStatusCode:
		responseModel, responseModelDiags := NewResourceTypeResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, resourceTypeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appDefinitionResourceTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceTypeModel

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

	response, err := r.providerData.client.DeleteResourceType(ctx, cm.DeleteResourceTypeParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
		ResourceTypeID:  state.ResourceTypeID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Resource type definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
