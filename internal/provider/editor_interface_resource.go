package provider

import (
	"context"
	"encoding/json"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                 = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithConfigure    = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithImportState  = (*editorInterfaceResource)(nil)
	_ resource.ResourceWithUpgradeState = (*editorInterfaceResource)(nil)
)

//nolint:ireturn
func NewEditorInterfaceResource() resource.Resource {
	return &editorInterfaceResource{}
}

type editorInterfaceResource struct {
	providerData ContentfulProviderData
}

func (r *editorInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_editor_interface"
}

func (r *editorInterfaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EditorInterfaceResourceSchema(ctx)
}

func (r *editorInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *editorInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *editorInterfaceResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// editorInterfaceResourceSchemaVersion0 := EditorInterfaceResourceSchemaForVersion(ctx, 0)
	// editorInterfaceResourceSchema := EditorInterfaceResourceSchema(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			// PriorSchema: &editorInterfaceResourceSchemaVersion0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {

				// type PriorEditorLayout struct {
				// 	Items []string `json:"items"`
				// }

				// type PriorState struct {
				// 	EditorLayout []PriorEditorLayout `json:"editor_layout"`
				// }

				var state EditorInterfaceResourceModelV0

				err := json.Unmarshal(req.RawState.JSON, &state)
				if err != nil {
					resp.Diagnostics.AddError(
						"Invalid Upgrade",
						"Failed to unmarshal state: "+err.Error(),
					)
					return
				}

				model, modelDiags := upgradeEditorInterfaceResourceModelV0ToV1(ctx, state)
				resp.Diagnostics.Append(modelDiags...)

				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, model)...)

				// resp.Diagnostics.AddError(
				// 	"Invalid Upgrade",
				// 	"NO",
				// )

				// var state EditorInterfaceResourceModel
				// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

				// stateEditorLayout := state.EditorLayout

				// var upgradedState EditorInterfaceResourceModel

				// var editorLayoutList []EditorInterfaceEditorLayoutElementValueV0
				// resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("editor_layout"), &editorLayoutList)...)

				// for i := 0; i < len(editorLayoutListElementValues); i++ {
				// 	var editorLayoutListElementValue EditorInterfaceEditorLayoutElementValueV0
				// 	sss, ok := editorLayoutListElementValues[i].(EditorInterfaceEditorLayoutElementValueV0)
				// 	if !ok {
				// 		resp.Diagnostics.AddAttributeError(
				// 			path.Root("editor_layout").AtListIndex(i),
				// 			"Invalid Upgrade",
				// 			"Editor layout element value is not of type EditorInterfaceEditorLayoutElementValueV0",
				// 		)
				// 		continue
				// 	}

				// 	editorLayoutListElementValueItemsElementValues := editorLayoutListElementValue.Items.Elements()
				// 	for j := 0; j < len(editorLayoutListElementValueItemsElementValues); j++ {
				// 		var editorLayoutListElementValueItemsElementValue jsontypes.Normalized
				// 		sss, ok := editorLayoutListElementValueItemsElementValues[j].(jsontypes.Normalized)
				// 		if !ok {
				// 			resp.Diagnostics.AddAttributeError(
				// 				path.Root("editor_layout").AtListIndex(i).AtListIndex(j),
				// 				"Invalid Upgrade",
				// 				"Editor layout element value items element value is not of type jsontypes.Normalized",
				// 			)
				// 			continue
				// 		}

				// 		// json.Unmarshal())
				// 		// var elementValue jsontypes.Normalized
				// 		// resp.Diagnostics.Append(editorLayoutListElementValueItemsElementValues[j].(ctx, j, &elementValue)...)
				// 		// if resp.Diagnostics.HasError() {
				// 		// 	continue
				// 		// }
				// 	}

				// 	// var elementValue EditorInterfaceEditorLayoutElementValueV0
				// 	// resp.Diagnostics.Append(editorLayoutListElementValues[i].(ctx, i, &elementValue)...)
				// 	// if resp.Diagnostics.HasError() {
				// 	// 	continue
				// 	// }
				// }

				// resp.State.SetAttribute(ctx, path.Root("editor_layout"), editorLayoutList)

				// resp.Diagnostics.AddAttributeError(
				// 	path.Root("editor_layout"),
				// 	"Invalid Upgrade",
				// 	"Editor layout cannot be upgraded from version 0 to 1",
				// )
			},
		},
	}
}

func (r *editorInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EditorInterfaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1
	currentVersion += r.providerData.editorInterfaceVersionOffset.Get(data.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToEditorInterfaceFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, params)

	tflog.Info(ctx, "editor_interface.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to create editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	r.providerData.editorInterfaceVersionOffset.Reset(data.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EditorInterfaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetEditorInterfaceParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ContentTypeID: data.ContentTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetEditorInterface(ctx, params)

	tflog.Info(ctx, "editor_interface.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *cm.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	r.providerData.editorInterfaceVersionOffset.Reset(data.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EditorInterfaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	currentVersion += r.providerData.editorInterfaceVersionOffset.Get(data.ContentTypeID.ValueString())

	params := cm.PutEditorInterfaceParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToEditorInterfaceFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEditorInterface(ctx, &request, params)

	tflog.Info(ctx, "editor_interface.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EditorInterface:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to update editor interface", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	r.providerData.editorInterfaceVersionOffset.Reset(data.ContentTypeID.ValueString())
}

func (r *editorInterfaceResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete editor interfaces
}
