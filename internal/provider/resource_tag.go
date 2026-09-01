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
	_ resource.Resource                = (*tagResource)(nil)
	_ resource.ResourceWithConfigure   = (*tagResource)(nil)
	_ resource.ResourceWithIdentity    = (*tagResource)(nil)
	_ resource.ResourceWithImportState = (*tagResource)(nil)
)

//nolint:ireturn
func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	providerData ContentfulProviderData
}

func tagIdentityAttributeNames() []string { return []string{"space_id", "environment_id", "tag_id"} }

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TagResourceSchema(ctx)
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *tagResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(tagIdentityAttributeNames())
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, tagIdentityAttributeNames(), req, resp)
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TagModel

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

	params, request, requestDiags := plan.ToPutTagRequest(path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutTag(ctx, &request, params)

	tflog.Info(ctx, "tag.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	version := 1

	var data TagModel

	switch response := response.(type) {
	case *cm.TagStatusCode:
		responseModel, responseModelDiags := NewTagResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create tag", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, tagIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagModel

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

	params := state.ToGetTagParams()

	response, err := r.providerData.client.GetTag(ctx, params)

	tflog.Info(ctx, "tag.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	version := 0

	var data TagModel

	switch response := response.(type) {
	case *cm.Tag:
		responseModel, responseModelDiags := NewTagResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read tag", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read tag", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, tagIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagModel

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

	params, request, requestDiags := plan.ToPutTagRequest(path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params.XContentfulVersion.SetTo(version)

	response, err := r.providerData.client.PutTag(ctx, &request, params)

	tflog.Info(ctx, "tag.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data TagModel

	switch response := response.(type) {
	case *cm.TagStatusCode:
		responseModel, responseModelDiags := NewTagResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update tag", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, tagIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagModel

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

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := state.ToDeleteTagParams()
	params.XContentfulVersion.SetTo(version)

	response, err := r.providerData.client.DeleteTag(ctx, params)

	tflog.Info(ctx, "tag.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Tag already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete tag", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
