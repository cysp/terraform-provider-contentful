package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ list.ListResource              = (*localeListResource)(nil)
	_ list.ListResourceWithConfigure = (*localeListResource)(nil)
)

//nolint:ireturn
func NewLocaleListResource() list.ListResource {
	return &localeListResource{}
}

type localeListResource struct {
	providerData ContentfulProviderData
}

func (r *localeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locale"
}

func (r *localeListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *localeListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = LocaleListResourceConfigSchema(ctx)
}

func (r *localeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config localeListResourceConfig

	configDiags := req.Config.Get(ctx, &config)
	if configDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(configDiags)

		return
	}

	spaceID, spaceIDDiags := requestRequiredString(config.SpaceID, path.Root("space_id"))
	environmentID, environmentIDDiags := requestRequiredString(config.EnvironmentID, path.Root("environment_id"))
	paramsDiags := diag.Diagnostics{}
	paramsDiags.Append(spaceIDDiags...)
	paramsDiags.Append(environmentIDDiags...)

	if paramsDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(paramsDiags)

		return
	}

	stream.Results = paginateContentfulCollectionItemsAsListResults(ctx, req,
		func(ctx context.Context, skip int64, limit int64) (contentfulCollection[cm.Locale], diag.Diagnostics) {
			const errorTitle = "Failed to list locales"

			pageParams := cm.GetLocalesParams{
				SpaceID:       spaceID,
				EnvironmentID: environmentID,
				Order:         []string{"sys.id"},
				Skip:          cm.NewOptInt64(skip),
				Limit:         cm.NewOptInt64(limit),
			}

			response, err := r.providerData.client.GetLocales(ctx, pageParams)

			tflog.Info(ctx, "locale.list", map[string]any{
				"params":   pageParams,
				"response": response,
				"err":      err,
			})

			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.LocaleCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(item cm.Locale) list.ListResult {
			model := NewLocaleResourceModelFromResponse(item)

			return newListResultFromResponse(
				ctx,
				req,
				item.Name,
				model.LocaleIdentityModel,
				func() (LocaleModel, diag.Diagnostics) {
					return model, nil
				},
			)
		},
	)
}
