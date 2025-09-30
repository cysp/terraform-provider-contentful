package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionResourceTypeResource)(nil)
	_ resource.ResourceWithMoveState   = (*appDefinitionResourceTypeResource)(nil)
)

//nolint:ireturn
func NewResourceTypeResource() resource.Resource {
	return &appDefinitionResourceTypeResource{}
}

type appDefinitionResourceTypeResource struct {
	providerData ContentfulProviderData
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
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
			"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
			"resource_type_id":  identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *appDefinitionResourceTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
		path.Root("resource_type_id"),
	}, req, resp)
}

func (r *appDefinitionResourceTypeResource) MoveState(ctx context.Context) []resource.StateMover {
	schema := ResourceTypeResourceSchema(ctx)

	return []resource.StateMover{
		{
			SourceSchema: &schema,
			StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				if req.SourceTypeName == "contentful_app_definition_resource_type" && req.SourceSchemaVersion == 0 {
					resp.TargetState = *req.SourceState

					return
				}
			},
		},
	}
}

//nolint:dupl
func (r *appDefinitionResourceTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	request, requestDiags := data.ToResourceTypeFields(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceType(ctx, &request, params)

	tflog.Info(ctx, "resource_type.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceTypeStatusCode:
		responseModel, responseModelDiags := NewResourceTypeResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create resource type definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel ResourceTypeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appDefinitionResourceTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetResourceType(ctx, params)

	tflog.Info(ctx, "resource_type.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

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

	var identityModel ResourceTypeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appDefinitionResourceTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
	}

	request, requestDiags := data.ToResourceTypeFields(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceType(ctx, &request, params)

	tflog.Info(ctx, "resource_type.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appDefinitionResourceTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteResourceType(ctx, cm.DeleteResourceTypeParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
		ResourceTypeID:  data.ResourceTypeID.ValueString(),
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
