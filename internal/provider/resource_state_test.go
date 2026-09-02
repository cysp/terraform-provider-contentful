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

type publicationBoolStateModel struct {
	ID types.Bool `tfsdk:"id"`
}

type publicationTwoStringStateModel struct {
	ID    types.String `tfsdk:"id"`
	Other types.String `tfsdk:"other"`
}

type invalidPublicationModel struct {
	Missing types.String `tfsdk:"missing"`
}

func TestSetResourceIdentityAndStateLeavesBothTargetsUnchangedAfterIdentityExtractionError(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		[]string{"missing"},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	require.True(t, diags.HasError())
	assertPublicationTargetsNull(t, identity, state)
}

func TestSetResourceIdentityAndStateLeavesBothTargetsUnchangedAfterStateEncodingError(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		[]string{"id"},
		&invalidPublicationModel{Missing: types.StringValue("invalid")},
	)

	require.True(t, diags.HasError())
	assertPublicationTargetsNull(t, identity, state)
}

func TestSetResourceIdentityAndStateSetsBothTargets(t *testing.T) {
	t.Parallel()

	identity, state := emptyPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		[]string{"id"},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	require.False(t, diags.HasError(), diags.Errors())

	var identityModel publicationIdentityModel

	var stateModel publicationStateModel

	require.False(t, identity.Get(t.Context(), &identityModel).HasError())
	require.False(t, state.Get(t.Context(), &stateModel).HasError())
	assert.Equal(t, types.StringValue("state"), identityModel.ID)
	assert.Equal(t, types.StringValue("state"), stateModel.ID)
}

func TestSetResourceIdentityAndStatePreservesNullAndUnknownIdentityValues(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			identity, state := emptyPublicationTargets()
			diags := setResourceIdentityAndState(
				t.Context(),
				identity,
				state,
				[]string{"id"},
				&publicationStateModel{ID: value},
			)

			require.False(t, diags.HasError(), diags.Errors())

			var identityModel publicationIdentityModel

			var stateModel publicationStateModel

			require.False(t, identity.Get(t.Context(), &identityModel).HasError())
			require.False(t, state.Get(t.Context(), &stateModel).HasError())
			assert.Equal(t, value, identityModel.ID)
			assert.Equal(t, value, stateModel.ID)
		})
	}
}

func TestSetResourceIdentityAndStateLeavesBothTargetsUnchangedForNonStringIdentityAttribute(t *testing.T) {
	t.Parallel()

	identity, state := nonStringPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		[]string{"id"},
		&publicationBoolStateModel{ID: types.BoolValue(true)},
	)

	require.True(t, diags.HasError())

	var identityModel publicationIdentityModel

	var stateModel publicationBoolStateModel

	require.False(t, identity.Get(t.Context(), &identityModel).HasError())
	require.False(t, state.Get(t.Context(), &stateModel).HasError())
	assert.True(t, identityModel.ID.IsNull())
	assert.True(t, stateModel.ID.IsNull())
}

func TestSetResourceIdentityAndStateLeavesBothTargetsUnchangedAfterLateIdentityEncodingError(t *testing.T) {
	t.Parallel()

	identity, state := lateIdentityEncodingErrorPublicationTargets()
	diags := setResourceIdentityAndState(
		t.Context(),
		identity,
		state,
		[]string{"id", "other"},
		&publicationTwoStringStateModel{
			ID:    types.StringValue("id"),
			Other: types.StringValue("other"),
		},
	)

	require.True(t, diags.HasError())

	var identityModel publicationIdentityModel

	var stateModel publicationTwoStringStateModel

	require.False(t, identity.Get(t.Context(), &identityModel).HasError())
	require.False(t, state.Get(t.Context(), &stateModel).HasError())
	assert.True(t, identityModel.ID.IsNull())
	assert.True(t, stateModel.ID.IsNull())
	assert.True(t, stateModel.Other.IsNull())
}

func TestSetResourceIdentityAndStateRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	_, state := emptyPublicationTargets()

	diags := setResourceIdentityAndState(
		t.Context(),
		nil,
		state,
		[]string{"id"},
		&publicationStateModel{ID: types.StringValue("state")},
	)

	assert.True(t, diags.HasError())
}

func nonStringPublicationTargets() (*tfsdk.ResourceIdentity, *tfsdk.State) {
	identityRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}
	stateRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.Bool}}

	return &tfsdk.ResourceIdentity{
		Schema: identityschema.Schema{Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		}},
		Raw: tftypes.NewValue(identityRawType, map[string]tftypes.Value{
			"id": tftypes.NewValue(tftypes.String, nil),
		}),
	}, &tfsdk.State{
		Schema: schema.Schema{Attributes: map[string]schema.Attribute{
			"id": schema.BoolAttribute{Computed: true},
		}},
		Raw: tftypes.NewValue(stateRawType, map[string]tftypes.Value{
			"id": tftypes.NewValue(tftypes.Bool, nil),
		}),
	}
}

func lateIdentityEncodingErrorPublicationTargets() (*tfsdk.ResourceIdentity, *tfsdk.State) {
	identityRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}
	stateRawType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":    tftypes.String,
		"other": tftypes.String,
	}}

	return &tfsdk.ResourceIdentity{
		Schema: identityschema.Schema{Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		}},
		Raw: tftypes.NewValue(identityRawType, map[string]tftypes.Value{
			"id": tftypes.NewValue(tftypes.String, nil),
		}),
	}, &tfsdk.State{
		Schema: schema.Schema{Attributes: map[string]schema.Attribute{
			"id":    schema.StringAttribute{Computed: true},
			"other": schema.StringAttribute{Computed: true},
		}},
		Raw: tftypes.NewValue(stateRawType, map[string]tftypes.Value{
			"id":    tftypes.NewValue(tftypes.String, nil),
			"other": tftypes.NewValue(tftypes.String, nil),
		}),
	}
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
