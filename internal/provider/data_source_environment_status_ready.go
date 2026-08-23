package provider

import (
	"context"
	"errors"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	environmentStatusReadyPollInterval = 15 * time.Second
	environmentStatusReadyTimeout      = 10 * time.Minute
	environmentStatusFailedValue       = "failed"
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
	var config EnvironmentStatusReadyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := config.Timeouts.Read(ctx, environmentStatusReadyTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	configuredTimeouts := config.Timeouts

	params := cm.GetEnvironmentParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(environmentStatusReadyPollInterval)
	defer ticker.Stop()

	for {
		response, err := d.providerData.client.GetEnvironment(ctx, params)

		tflog.Info(ctx, "environment_status_ready.read", map[string]any{
			"params":   params,
			"response": response,
			"err":      err,
		})

		switch response := response.(type) {
		case *cm.Environment:
			continuePolling, responseDiags := publishEnvironmentStatusReadyResponse(
				ctx,
				&resp.State,
				*response,
				configuredTimeouts,
			)
			resp.Diagnostics.Append(responseDiags...)

			if resp.Diagnostics.HasError() || !continuePolling {
				return
			}

		default:
			resp.Diagnostics.AddError("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				resp.Diagnostics.AddError(
					"Timed out waiting for environment to become ready",
					"",
				)
			}

			return
		case <-ticker.C:
		}
	}
}
