package provider

import (
	"context"
	"net/http"
	"os"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Factory(version string, options ...Option) func() provider.Provider {
	return func() provider.Provider {
		return New(version, options...)
	}
}

func New(version string, options ...Option) *ContentfulProvider {
	provider := ContentfulProvider{
		version: version,
	}

	for _, option := range options {
		option(&provider)
	}

	return &provider
}

type ContentfulProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	contentfulURL string
	httpClient    *http.Client
	accessToken   string
}

type ContentfulProviderModel struct {
	URL         types.String `tfsdk:"url"`
	AccessToken types.String `tfsdk:"access_token"`
}

var _ provider.Provider = (*ContentfulProvider)(nil)

type Option func(*ContentfulProvider)

func WithContentfulURL(url string) Option {
	return func(p *ContentfulProvider) {
		p.contentfulURL = url
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(p *ContentfulProvider) {
		p.httpClient = httpClient
	}
}

func WithAccessToken(accessToken string) Option {
	return func(p *ContentfulProvider) {
		p.accessToken = accessToken
	}
}

func (p *ContentfulProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Contentful space configuration.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
			},
			"access_token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *ContentfulProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ContentfulProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var contentfulURL string
	if !data.URL.IsNull() {
		contentfulURL = data.URL.ValueString()
	} else if contentfulURLFromEnv, found := os.LookupEnv("CONTENTFUL_URL"); found {
		contentfulURL = contentfulURLFromEnv
	}

	if p.contentfulURL != "" {
		contentfulURL = p.contentfulURL
	}

	if contentfulURL == "" {
		contentfulURL = cm.DefaultServerURL
	}

	if contentfulURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Failed to configure client", "No API URL provided")
	}

	var accessToken string
	if !data.AccessToken.IsNull() {
		accessToken = data.AccessToken.ValueString()
	} else {
		if accessTokenFromEnv, found := os.LookupEnv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN"); found {
			accessToken = accessTokenFromEnv
		}
	}

	if p.accessToken != "" {
		accessToken = p.accessToken
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(path.Root("access_token"), "Failed to configure client", "No access token provided")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryWaitMin = time.Duration(1) * time.Second
	retryableClient.RetryWaitMax = time.Duration(3) * time.Second //nolint:mnd
	retryableClient.Backoff = retryablehttp.LinearJitterBackoff

	if p.httpClient != nil {
		retryableClient.HTTPClient = p.httpClient
	}

	contentfulManagementClient, err := cm.NewClient(
		contentfulURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(util.NewClientWithUserAgent(retryableClient.StandardClient(), "terraform-provider-contentful/"+p.version)),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Contentful client: %s", err.Error())
	}

	providerData := ContentfulProviderData{
		client:                       contentfulManagementClient,
		editorInterfaceVersionOffset: &ContentfulContentTypeCounter{},
	}

	resp.ActionData = providerData
	resp.DataSourceData = providerData
	resp.ListResourceData = providerData
	resp.ResourceData = providerData
}

func (p *ContentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contentful"
	resp.Version = p.version
}

func (p *ContentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAppDefinitionDataSource,
		NewMarketplaceAppDefinitionDataSource,
		NewPreviewAPIKeyDataSource,
	}
}

func (p *ContentfulProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		NewContentTypeListResource,
		NewEntryListResource,
	}
}

func (p *ContentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppDefinitionResource,
		NewAppSigningSecretResource,
		NewAppInstallationResource,
		NewContentTypeResource,
		NewDeliveryAPIKeyResource,
		NewEditorInterfaceResource,
		NewEnvironmentAliasResource,
		NewEnvironmentResource,
		NewEntryResource,
		NewExtensionResource,
		NewPersonalAccessTokenResource,
		NewResourceProviderResource,
		NewResourceTypeResource,
		NewRoleResource,
		NewSpaceEnablementsResource,
		NewTagResource,
		NewTeamResource,
		NewTeamSpaceMembershipResource,
		NewWebhookResource,
	}
}
