//nolint:testpackage // Resource state publication is intentionally tested through its package-private seam.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publicationIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

type publicationStateModel struct {
	ID types.String `tfsdk:"id"`
}

type invalidPublicationModel struct {
	Missing types.String `tfsdk:"missing"`
}

func TestSetResourceIdentityAndStateIsAtomicAfterIdentityError(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		&invalidPublicationModel{Missing: types.StringValue("invalid")},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	require.True(t, diags.HasError())
	assertPublicationTargetsNull(t, identity, state)
}

func TestSetResourceIdentityAndStateIsAtomicAfterStateError(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		&publicationIdentityModel{ID: types.StringValue("identity")},
		&invalidPublicationModel{Missing: types.StringValue("invalid")},
	)

	require.True(t, diags.HasError())
	assertPublicationTargetsNull(t, identity, state)
}

func TestSetResourceIdentityAndStatePublishesBoth(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		&publicationIdentityModel{ID: types.StringValue("identity")},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	require.False(t, diags.HasError(), diags.Errors())

	var identityModel publicationIdentityModel

	var stateModel publicationStateModel

	require.False(t, identity.Get(t.Context(), &identityModel).HasError())
	require.False(t, state.Get(t.Context(), &stateModel).HasError())
	assert.Equal(t, types.StringValue("identity"), identityModel.ID)
	assert.Equal(t, types.StringValue("state"), stateModel.ID)
}

func TestSetResourceIdentityAndStateRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	_, state := emptyPublicationTargets()

	diags := setResourceIdentityAndState(
		t.Context(),
		nil,
		state,
		&publicationIdentityModel{ID: types.StringValue("identity")},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	assert.True(t, diags.HasError())
}

func emptyPublicationTargets() (*tfsdk.ResourceIdentity, *tfsdk.State) {
	rawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}
	raw := tftypes.NewValue(rawType, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, nil),
	})

	return &tfsdk.ResourceIdentity{
			Schema: identityschema.Schema{Attributes: map[string]identityschema.Attribute{
				"id": identityschema.StringAttribute{RequiredForImport: true},
			}},
			Raw: raw,
		}, &tfsdk.State{
			Schema: schema.Schema{Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true},
			}},
			Raw: raw,
		}
}

func assertPublicationTargetsNull(t *testing.T, identity *tfsdk.ResourceIdentity, state *tfsdk.State) {
	t.Helper()

	var identityModel publicationIdentityModel

	var stateModel publicationStateModel

	require.False(t, identity.Get(t.Context(), &identityModel).HasError())
	require.False(t, state.Get(t.Context(), &stateModel).HasError())
	assert.True(t, identityModel.ID.IsNull())
	assert.True(t, stateModel.ID.IsNull())
}
