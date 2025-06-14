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
	_ resource.Resource                = (*spaceEnablementsResource)(nil)
	_ resource.ResourceWithConfigure   = (*spaceEnablementsResource)(nil)
	_ resource.ResourceWithImportState = (*spaceEnablementsResource)(nil)
)

//nolint:ireturn
func NewSpaceEnablementsResource() resource.Resource {
	return &spaceEnablementsResource{}
}

type spaceEnablementsResource struct {
	providerData ContentfulProviderData
}

func (r *spaceEnablementsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_enablements"
}

func (r *spaceEnablementsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SpaceEnablementsResourceSchema(ctx)
}

func (r *spaceEnablementsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *spaceEnablementsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("space_id"), req, resp)
}

func (r *spaceEnablementsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SpaceEnablementsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := cm.PutSpaceEnablementsParams{
		SpaceID:            data.SpaceID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToSpaceEnablementFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutSpaceEnablements(ctx, &request, params)

	tflog.Info(ctx, "space_enablements.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.SpaceEnablementStatusCode:
		responseModel, responseModelDiags := NewSpaceEnablementsResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spaceEnablementsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SpaceEnablementsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetSpaceEnablementsParams{
		SpaceID: data.SpaceID.ValueString(),
	}

	response, err := r.providerData.client.GetSpaceEnablements(ctx, params)

	tflog.Info(ctx, "space_enablements.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *cm.GetSpaceEnablementsApplicationJSONOK:
		responseModel, responseModelDiags := NewSpaceEnablementsResourceModelFromResponse(ctx, cm.SpaceEnablement(*response))
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	case *cm.GetSpaceEnablementsApplicationVndContentfulManagementV1JSONOK:
		responseModel, responseModelDiags := NewSpaceEnablementsResourceModelFromResponse(ctx, cm.SpaceEnablement(*response))
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spaceEnablementsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SpaceEnablementsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := cm.PutSpaceEnablementsParams{
		SpaceID:            data.SpaceID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToSpaceEnablementFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutSpaceEnablements(ctx, &request, params)

	tflog.Info(ctx, "space_enablements.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.SpaceEnablementStatusCode:
		responseModel, responseModelDiags := NewSpaceEnablementsResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spaceEnablementsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete space enablements
}
