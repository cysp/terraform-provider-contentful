package provider

import (
	"context"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	environmentStatusReadyPollInterval = 1 * time.Second
	environmentStatusReadyTimeout      = 10 * time.Minute
	environmentStatusReadyValue        = "ready"
)

var (
	_ datasource.DataSource              = (*environmentStatusReadyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentStatusReadyDataSource)(nil)
)

//nolint:ireturn
func NewEnvironmentStatusReadyDataSource() datasource.DataSource {
	return &environmentStatusReadyDataSource{}
}

type environmentStatusReadyDataSource struct {
	providerData ContentfulProviderData
}

func (d *environmentStatusReadyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_status_ready"
}

func (d *environmentStatusReadyDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvironmentStatusReadyDataSourceSchema(ctx)
}

func (d *environmentStatusReadyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *environmentStatusReadyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentStatusReadyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetEnvironmentParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, environmentStatusReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(environmentStatusReadyPollInterval)
	defer ticker.Stop()

	for {
		response, err := d.providerData.client.GetEnvironment(ctx, params)

		tflog.Debug(ctx, "environment_status_ready.read", map[string]any{
			"params":   params,
			"response": response,
			"err":      err,
		})

		var data EnvironmentStatusReadyModel

		switch response := response.(type) {
		case *cm.Environment:
			responseModel, responseModelDiags := NewEnvironmentStatusReadyModelFromResponse(ctx, *response)
			resp.Diagnostics.Append(responseModelDiags...)

			data = responseModel

		default:
			resp.Diagnostics.AddError("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		if data.Status.ValueString() == environmentStatusReadyValue {
			return
		}

		select {
		case <-ctx.Done():

			return
		case <-ctxWithTimeout.Done():
			resp.Diagnostics.AddError(
				"Timed out waiting for environment to become ready",
				"",
			)

			return
		case <-ticker.C:
		}
	}
}
