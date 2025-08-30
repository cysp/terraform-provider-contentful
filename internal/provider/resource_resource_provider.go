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
	_ resource.Resource                = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithMoveState   = (*appDefinitionResourceProviderResource)(nil)
)

//nolint:ireturn
func NewResourceProviderResource() resource.Resource {
	return &appDefinitionResourceProviderResource{}
}

type appDefinitionResourceProviderResource struct {
	providerData ContentfulProviderData
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
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
			"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *appDefinitionResourceProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

func (r *appDefinitionResourceProviderResource) MoveState(ctx context.Context) []resource.StateMover {
	schema := ResourceProviderResourceSchema(ctx)

	return []resource.StateMover{
		{
			SourceSchema: &schema,
			StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				if req.SourceTypeName == "contentful_app_definition_resource_provider" && req.SourceSchemaVersion == 0 {
					resp.TargetState = *req.SourceState

					return
				}
			},
		},
	}
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	request, requestDiags := data.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "resource_provider.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetResourceProvider(ctx, params)

	tflog.Info(ctx, "resource_provider.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceProvider:
		responseModel, responseModelDiags := NewResourceProviderResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	request, requestDiags := data.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "resource_provider.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionResourceProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteResourceProvider(ctx, cm.DeleteResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Resource provider definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
