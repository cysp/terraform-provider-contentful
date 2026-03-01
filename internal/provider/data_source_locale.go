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
	_ datasource.DataSource              = (*localeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*localeDataSource)(nil)
)

//nolint:ireturn
func NewLocaleDataSource() datasource.DataSource {
	return &localeDataSource{}
}

type localeDataSource struct {
	providerData ContentfulProviderData
}

func (d *localeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locale"
}

func (d *localeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = LocaleDataSourceSchema(ctx)
}

func (d *localeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *localeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LocaleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetLocaleParams{
		SpaceID:  data.SpaceID.ValueString(),
		LocaleID: data.LocaleID.ValueString(),
	}

	response, err := d.providerData.client.GetLocale(ctx, params)

	tflog.Info(ctx, "locale_data_source.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.Locale:
		responseModel, responseModelDiags := NewLocaleDataSourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
