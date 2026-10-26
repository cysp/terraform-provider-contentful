package provider

// This tests the private staged publication boundary after endpoint correction.
import (
	"testing"

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
