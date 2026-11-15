package provider

// This tests the private staged publication boundary after endpoint correction.
import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutationRecoveryEncodingFailureLeavesStateAndIdentityUnchanged(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	stateSchema := TeamResourceSchema(ctx)
	state := tfsdk.State{Schema: stateSchema, Raw: tftypes.NewValue(stateSchema.Type().TerraformType(ctx), nil)}
	identitySchema := identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"organization_id": identityschema.StringAttribute{RequiredForImport: true},
		"team_id":         identityschema.BoolAttribute{RequiredForImport: true},
	}}
	identity := tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil)}
	originalState, originalIdentity := state.Raw, identity.Raw
	response := TeamModel{
		ID: types.StringValue("other-org/returned-team"), OrganizationID: types.StringValue("other-org"), TeamID: types.StringValue("returned-team"),
		Name: types.StringValue("Returned"), Description: types.StringNull(), Timeouts: TimeoutsNull(),
	}
	publicationDiags, identityDiags := setResourceMutationState(ctx, &identity, &state, []string{"organization_id", "team_id"}, &response, "team", map[string]string{"organization_id": "org"})
	require.True(t, publicationDiags.HasError())
	require.True(t, identityDiags.HasError())
	assert.True(t, originalState.Equal(state.Raw))
	assert.True(t, originalIdentity.Equal(identity.Raw))
	assert.Equal(t, types.StringValue("other-org"), response.OrganizationID)
	assert.Equal(t, types.StringValue("other-org/returned-team"), response.ID)
}

func TestMutationVersionPublicationPreservesFailureBoundaries(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		identityConfirmed bool
		encodingFails     bool
		privateFails      bool
		expectedVersion   string
	}{
		"confirmed version":                    {identityConfirmed: true, expectedVersion: "42"},
		"revoked version":                      {},
		"failed version write":                 {identityConfirmed: true, privateFails: true, expectedVersion: "17"},
		"failed revocation":                    {privateFails: true, expectedVersion: "17"},
		"encoding failure precedes revocation": {encodingFails: true, expectedVersion: "17"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			identity, state := emptyPublicationTargets()

			attributes := []string{"id"}
			if test.encodingFails {
				attributes = []string{"missing"}
			}

			writes := 0
			private := &publicationPrivateData{data: map[string][]byte{"version": []byte("17")}, write: func(key string, _ []byte) diag.Diagnostics {
				writes++

				assert.Equal(t, "version", key)

				var model publicationStateModel
				require.Empty(t, state.Get(t.Context(), &model))
				assert.Equal(t, types.StringValue("new"), model.ID)

				if test.privateFails {
					return diag.Diagnostics{diag.NewErrorDiagnostic("Private write failed", "The test storage rejected the version change.")}
				}

				return nil
			}}
			diags := setResourceIdentityStateAndMutationVersion(t.Context(), identity, state, private, attributes, &publicationStateModel{ID: types.StringValue("new")}, 42, test.identityConfirmed)
			assert.Equal(t, test.encodingFails || test.privateFails, diags.HasError())
			assert.Equal(t, test.expectedVersion, string(private.data["version"]))

			if test.encodingFails {
				assert.Zero(t, writes)
				assertPublicationTargetsNull(t, identity, state)
			} else {
				assert.Equal(t, 1, writes)

				var model publicationIdentityModel
				require.Empty(t, identity.Get(t.Context(), &model))
				assert.Equal(t, types.StringValue("new"), model.ID)
			}
		})
	}
}
