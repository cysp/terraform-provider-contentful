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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	httpClient *http.Client
}

var _ provider.Provider = (*ContentfulProvider)(nil)

type Option func(*ContentfulProvider)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(p *ContentfulProvider) {
		p.httpClient = httpClient
	}
}

func (p *ContentfulProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = ContentfulProviderSchema(ctx)
	resp.Schema.Description = "Manage Contentful space configuration."
}

func (p *ContentfulProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ContentfulModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var contentfulURL string
	if !data.Url.IsNull() {
		contentfulURL = data.Url.ValueString()
	} else if contentfulURLFromEnv, found := os.LookupEnv("CONTENTFUL_URL"); found {
		contentfulURL = contentfulURLFromEnv
	} else {
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

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *ContentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contentful"
	resp.Version = p.version
}

func (p *ContentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPreviewApiKeyDataSource,
	}
}

func (p *ContentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppInstallationResource,
		NewContentTypeResource,
		NewDeliveryApiKeyResource,
		NewEditorInterfaceResource,
		NewPersonalAccessTokenResource,
		NewRoleResource,
		NewWebhookResource,
	}
}
