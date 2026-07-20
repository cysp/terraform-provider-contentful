package provider

import (
	"errors"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInjectedTeamListTransport = errors.New("injected transport failure")

func TestTeamsDataSourceTeamSchemaMatchesTeamResource(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	dataSourceSchema := TeamsDataSourceSchema(ctx)
	teamsAttribute, ok := dataSourceSchema.Attributes["teams"].(datasourceschema.ListNestedAttribute)
	require.True(t, ok)

	itemAttributes := teamsAttribute.NestedObject.Attributes
	require.Len(t, itemAttributes, 3)
	assert.NotContains(t, itemAttributes, "id")
	assert.NotContains(t, itemAttributes, "organization_id")

	resourceAttributes := TeamResourceSchema(ctx).Attributes

	for _, name := range []string{"team_id", "name", "description"} {
		itemAttribute, ok := itemAttributes[name]
		require.True(t, ok, "missing teams item attribute %q", name)
		resourceAttribute, ok := resourceAttributes[name]
		require.True(t, ok, "missing team resource attribute %q", name)

		assert.Equal(t, types.StringType, itemAttribute.GetType())
		assert.True(t, itemAttribute.IsComputed())
		assert.False(t, itemAttribute.IsOptional())
		assert.False(t, itemAttribute.IsRequired())
		assert.Equal(t, resourceAttribute.GetType(), itemAttribute.GetType())
		assert.Equal(t, resourceAttribute.GetDescription(), itemAttribute.GetDescription())
	}
}

func TestReadTeamsTransportError(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errInjectedTeamListTransport
		}),
	}

	client, err := cm.NewClient(
		"https://api.contentful.invalid",
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(cm.NewTransportClient(httpClient, "test")),
	)
	require.NoError(t, err)

	teams, diagnostics := readTeams(t.Context(), client, "organization-id")

	assert.Nil(t, teams)
	require.Len(t, diagnostics, 1)
	assert.Equal(t, "Failed to read teams", diagnostics[0].Summary())
	assert.Contains(t, diagnostics[0].Detail(), "injected transport failure")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
