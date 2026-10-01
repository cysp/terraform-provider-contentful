package provider

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed path-bound vector, computed independently of the provider helper with
// salt bytes 0..15, value 64 ASCII a's, and the documented Argon2id parameters.
const writeOnlyTestHash = "$argon2id$v=19$m=65536,t=1,p=4$AAECAwQFBgcICQoLDA0ODw$8fzTBn/c7LLcUcsOYeZ+mTkErHvsHBw+y3yis5C8IQk"

type writeOnlyTestPrivateData struct {
	data     map[string][]byte
	fail     bool
	failRead bool
	writes   int
}

func (p *writeOnlyTestPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if p.failRead {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Private read failed", "Test failure")}
	}

	return p.data[key], nil
}

func (p *writeOnlyTestPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	p.writes++
	if p.fail {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Private write failed", "Test failure")}
	}

	p.data[key] = bytes.Clone(value)

	return nil
}

// Verify the previously emitted singleton encoding and its independent Argon2
// vector. A matching value must preserve the exact bytes, including its salt.
func TestSigningSecretReadsAcknowledgedVerifier(t *testing.T) {
	t.Parallel()

	data := []byte(testJSON([]map[string]string{{"path": "value_wo", "hash": writeOnlyTestHash}}))
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
	verifier, diags := readSigningSecretVerifier(t.Context(), private, types.StringNull())
	require.Empty(t, diags)
	matched, diags := verifier.matches(t.Context(), types.StringValue(strings.Repeat("a", 64)))
	require.Empty(t, diags)
	require.True(t, matched)
	require.Empty(t, verifier.publish(t.Context(), private))
	assert.Zero(t, private.writes)
	assert.Equal(t, data, private.data[writeOnlySecretHashesPrivateKey])
}

func TestWriteOnlySecretRejectsInvalidPrivateData(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	cases := map[string]struct {
		data   []byte
		reason string
	}{
		"null":                   {data: []byte(testJSON(nil)), reason: "non-null JSON array"},
		"object":                 {data: []byte(testJSON(map[string]any{})), reason: "non-null JSON array"},
		"truncated":              {data: []byte(`[`), reason: "non-null JSON array"},
		"trailing":               {data: []byte(writeOnlyTestPrivate + `[]`), reason: "Unexpected data follows"},
		"unknown field":          {data: []byte(testJSON([]map[string]any{{"path": "value_wo", "hash": writeOnlyTestHash, "unexpected": 42}})), reason: "no additional fields"},
		"empty path":             {data: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "", Hash: writeOnlyTestHash}})), reason: "not supported by this signing-secret resource"},
		"duplicate path":         {data: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}, {Path: "value_wo", Hash: writeOnlyTestHash}})), reason: "at most one signing secret verifier"},
		"unsupported algorithm":  {data: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "argon2id", "argon2i", 1)}})), reason: "unsupported algorithm, version, or parameter encoding"},
		"unsupported parameters": {data: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "m=65536", "m=4294967295", 1)}})), reason: "unsupported algorithm, version, or parameter encoding"},
		"invalid key":            {data: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "8fzTBn/", "!!!!!!!", 1)}})), reason: "invalid salt or key encoding or length"},
		"excessive size":         {data: bytes.Repeat([]byte(" "), writeOnlySecretHashesMaxLength+1), reason: "exceeds the size limit"},
		"invalid UTF8":           {data: append([]byte(writeOnlyTestPrivate), 0xff), reason: "not valid UTF-8"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: test.data}}
			_, diags := readSigningSecretVerifier(t.Context(), private, types.StringNull())
			require.True(t, diags.HasError())
			assert.Equal(t, "Invalid write-only secret private state", diags.Errors()[0].Summary())
			assert.Contains(t, diags.Errors()[0].Detail(), test.reason)
			assert.NotContains(t, diags.Errors()[0].Detail(), writeOnlyTestHash)
			assert.Zero(t, private.writes)
		})
	}
}

func TestWriteOnlySecretMissingIsNotCorrupt(t *testing.T) {
	t.Parallel()

	for _, data := range [][]byte{nil, []byte(testJSON([]any{}))} {
		private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
		verifier, diags := readSigningSecretVerifier(t.Context(), private, types.StringNull())
		require.Empty(t, diags)
		matched, diags := verifier.matches(t.Context(), types.StringValue(strings.Repeat("a", 64)))
		require.Empty(t, diags)
		assert.False(t, matched)
	}
}

func TestWriteOnlySecretCancellationDoesNotPrepareOrPublish(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	verifier, diags := readSigningSecretVerifier(t.Context(), private, types.StringNull())
	require.Empty(t, diags)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.True(t, verifier.set(ctx, types.StringValue(strings.Repeat("b", 64))).HasError())
	assert.JSONEq(t, writeOnlyTestPrivate, string(verifier.prepared))
	assert.Zero(t, private.writes)
}

func TestSigningSecretPublicationFailurePreservesAcknowledgedState(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	for _, test := range []struct {
		name              string
		failPrivateWrite  bool
		invalidModel      bool
		discardObsolete   bool
		wantPrivateWrites int
	}{
		{name: "private write", failPrivateWrite: true, wantPrivateWrites: 1},
		{name: "public state encoding", invalidModel: true},
		{name: "obsolete verifier private write", discardObsolete: true, failPrivateWrite: true, wantPrivateWrites: 1},
		{name: "obsolete verifier public state encoding", discardObsolete: true, invalidModel: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			prior := appSigningSecretTestModel(strings.Repeat("a", 64))
			state := tfsdk.State{Schema: AppSigningSecretResourceSchema(ctx)}
			identity := appSigningSecretTestIdentity(ctx)
			require.Empty(t, setResourceIdentityAndState(ctx, identity, &state, appSigningSecretIdentityAttributeNames(), &prior))
			priorState, priorIdentity := state.Raw.Copy(), identity.Raw.Copy()

			privateData := writeOnlyTestPrivate
			if test.discardObsolete {
				privateData = "invalid verifier data"
			}

			private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(privateData)}, fail: test.failPrivateWrite}

			var verifier signingSecretVerifier

			var diags diag.Diagnostics
			if test.discardObsolete {
				verifier, diags = readSigningSecretVerifier(ctx, private, prior.Value)
				require.Empty(t, diags)
			} else {
				verifier, diags = readSigningSecretVerifier(ctx, private, types.StringNull())
				require.Empty(t, diags)
				require.Empty(t, verifier.set(ctx, types.StringValue(strings.Repeat("b", 64))))
			}

			changed := appSigningSecretTestModel(strings.Repeat("b", 64))
			changed.AppDefinitionID = types.StringValue("different-app")

			var model any = &changed
			if test.invalidModel {
				model = "cannot encode a string as resource state"
			}

			diags = publishSigningSecretState(ctx, private, &verifier, identity, &state, appSigningSecretIdentityAttributeNames(), model)
			require.True(t, diags.HasError())
			assert.True(t, priorState.Equal(state.Raw))
			assert.True(t, priorIdentity.Equal(identity.Raw))
			assert.Equal(t, privateData, string(private.data[writeOnlySecretHashesPrivateKey]))

			assert.Equal(t, test.wantPrivateWrites, private.writes)

			if test.failPrivateWrite {
				assert.Equal(t, "Private write failed", diags.Errors()[0].Summary())
			}
		})
	}
}

func TestSigningSecretOrdinaryValueRequiresPrivateAccess(t *testing.T) {
	t.Parallel()

	private := &writeOnlyTestPrivateData{failRead: true}
	_, diags := readSigningSecretVerifier(t.Context(), private, types.StringValue(strings.Repeat("a", 64)))
	require.Len(t, diags, 1)
	assert.Equal(t, "Private read failed", diags[0].Summary())
	assert.Zero(t, private.writes)
}

func TestSigningSecretRejectsOtherVerifierPaths(t *testing.T) {
	t.Parallel()

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "other", Hash: writeOnlyTestHash}}))}}
	_, diags := readSigningSecretVerifier(t.Context(), private, types.StringNull())
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "not supported by this signing-secret resource")
	assert.NotContains(t, diags.Errors()[0].Detail(), writeOnlyTestHash)
	assert.Zero(t, private.writes)
}
