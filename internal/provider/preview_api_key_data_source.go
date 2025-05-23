package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = (*previewApiKeyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*previewApiKeyDataSource)(nil)
)

//nolint:ireturn,revive
func NewPreviewApiKeyDataSource() datasource.DataSource {
	return &previewApiKeyDataSource{}
}

//nolint:revive
type previewApiKeyDataSource struct {
	providerData ContentfulProviderData
}

func (d *previewApiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_preview_api_key"
}

func (d *previewApiKeyDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PreviewAPIKeyDataSourceSchema(ctx)
}

func (d *previewApiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

//nolint:dupl
func (d *previewApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PreviewAPIKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetPreviewApiKeyParams{
		SpaceID:         data.SpaceID.ValueString(),
		PreviewAPIKeyID: data.PreviewAPIKeyID.ValueString(),
	}

	response, err := d.providerData.client.GetPreviewApiKey(ctx, params)

	tflog.Info(ctx, "preview_api_key.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.PreviewApiKey:
		responseModel, responseModelDiags := NewPreviewAPIKeyDataSourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read preview api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read preview api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
