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
	_ resource.Resource                = (*appDefinitionAppDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionAppDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionAppDefinitionResourceTypeResource)(nil)
)

//nolint:ireturn
func NewAppDefinitionResourceTypeResource() resource.Resource {
	return &appDefinitionAppDefinitionResourceTypeResource{}
}

type appDefinitionAppDefinitionResourceTypeResource struct {
	providerData ContentfulProviderData
}

func (r *appDefinitionAppDefinitionResourceTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_definition_resource_type"
}

func (r *appDefinitionAppDefinitionResourceTypeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppDefinitionResourceTypeResourceSchema(ctx)
}

func (r *appDefinitionAppDefinitionResourceTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionAppDefinitionResourceTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
		path.Root("resource_type_id"),
	}, req, resp)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppDefinitionResourceTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutAppDefinitionResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	request, requestDiags := data.ToResourceTypeFields(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppDefinitionResourceType(ctx, &request, params)

	tflog.Info(ctx, "app_definition_resource_type.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceTypeStatusCode:
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response.Response)...)

	default:
		resp.Diagnostics.AddError("Failed to create resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appDefinitionAppDefinitionResourceTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppDefinitionResourceTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetAppDefinitionResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetAppDefinitionResourceType(ctx, params)

	tflog.Info(ctx, "app_definition_resource_type.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceType:
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, *response)...)

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppDefinitionResourceTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutAppDefinitionResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	request, requestDiags := data.ToResourceTypeFields(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppDefinitionResourceType(ctx, &request, params)

	tflog.Info(ctx, "app_definition_resource_type.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceTypeStatusCode:
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response.Response)...)

	default:
		resp.Diagnostics.AddError("Failed to update resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appDefinitionAppDefinitionResourceTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppDefinitionResourceTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteAppDefinitionResourceType(ctx, cm.DeleteAppDefinitionResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Resource type definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
