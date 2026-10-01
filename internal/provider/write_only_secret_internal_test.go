package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed path-bound vector, computed independently of the provider helper with
// salt bytes 0..15, value 64 ASCII a's, and the documented Argon2id parameters.
const writeOnlyTestHash = "$argon2id$v=19$m=65536,t=1,p=4$AAECAwQFBgcICQoLDA0ODw$8fzTBn/c7LLcUcsOYeZ+mTkErHvsHBw+y3yis5C8IQk"

type writeOnlyTestPrivateData struct {
	data   map[string][]byte
	fail   bool
	writes int
}

func (p *writeOnlyTestPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
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

func TestWriteOnlySecretVerifier(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})
	ctx := t.Context()
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate), "unrelated": []byte(testJSON(42))}}
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)

	matches, diags := hashes.matches(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("a", 64)))
	require.Empty(t, diags)
	assert.True(t, matches)
	matches, diags = hashes.matches(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("b", 64)))
	require.Empty(t, diags)
	assert.False(t, matches)

	unchanged, diags := prepareSigningSecretValue(ctx, &hashes, types.StringNull(), types.StringValue(strings.Repeat("a", 64)), path.Root("value_wo"))
	require.Empty(t, diags)
	assert.True(t, unchanged)
	require.Empty(t, hashes.publish(ctx, private))
	assert.Zero(t, private.writes)
	assert.JSONEq(t, writeOnlyTestPrivate, string(private.data[writeOnlySecretHashesPrivateKey]))

	require.Empty(t, hashes.set(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("b", 64))))
	assert.JSONEq(t, writeOnlyTestPrivate, string(private.data[writeOnlySecretHashesPrivateKey]), "preparation must not acknowledge a value")
	require.Empty(t, hashes.publish(ctx, private))
	assert.NotContains(t, string(private.data[writeOnlySecretHashesPrivateKey]), strings.Repeat("b", 64))
	assert.Equal(t, []byte(testJSON(42)), private.data["unrelated"])
	assert.Equal(t, 1, private.writes)

	reloaded, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)
	matches, diags = reloaded.matches(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("b", 64)))
	require.Empty(t, diags)
	assert.True(t, matches)

	require.Empty(t, reloaded.remove(path.Root("value_wo")))
	require.Empty(t, reloaded.publish(ctx, private))
	assert.Empty(t, private.data[writeOnlySecretHashesPrivateKey])
	assert.Equal(t, []byte(testJSON(42)), private.data["unrelated"])
}

func TestWriteOnlySecretMultiplePaths(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})
	ctx := t.Context()
	root := path.Root("value_wo")
	nested := path.Root("credentials").AtName("token_wo")
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)
	require.Empty(t, hashes.set(ctx, nested, types.StringValue("nested-secret")))
	require.Empty(t, hashes.publish(ctx, private))

	var records []writeOnlySecretHashRecord
	require.NoError(t, json.Unmarshal(private.data[writeOnlySecretHashesPrivateKey], &records))
	require.Len(t, records, 2)
	assert.Equal(t, "credentials.token_wo", records[0].Path)
	assert.Equal(t, "value_wo", records[1].Path)
	assert.Equal(t, writeOnlyTestHash, records[1].Hash)
	nestedHash := records[0].Hash

	reloaded, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)
	matched, diags := reloaded.matches(ctx, nested, types.StringValue("nested-secret"))
	require.Empty(t, diags)
	assert.True(t, matched)
	matched, diags = reloaded.matches(ctx, root, types.StringValue(strings.Repeat("a", 64)))
	require.Empty(t, diags)
	assert.True(t, matched)
	require.Empty(t, reloaded.set(ctx, root, types.StringValue("rotated-root-secret")))
	require.Empty(t, reloaded.publish(ctx, private))
	require.NoError(t, json.Unmarshal(private.data[writeOnlySecretHashesPrivateKey], &records))
	assert.Equal(t, nestedHash, records[0].Hash)
	require.Empty(t, reloaded.remove(root))
	require.Empty(t, reloaded.publish(ctx, private))
	require.NoError(t, json.Unmarshal(private.data[writeOnlySecretHashesPrivateKey], &records))
	require.Len(t, records, 1)
	assert.Equal(t, "credentials.token_wo", records[0].Path)
	assert.Equal(t, nestedHash, records[0].Hash)
	require.Empty(t, reloaded.remove(nested))
	require.Empty(t, reloaded.publish(ctx, private))
	assert.Empty(t, private.data[writeOnlySecretHashesPrivateKey])
}

func TestWriteOnlySecretVerifierIsBoundToPath(t *testing.T) {
	t.Parallel()

	nested := path.Root("credentials").AtName("token_wo")
	transplanted := testJSON([]writeOnlySecretHashRecord{{Path: "credentials.token_wo", Hash: writeOnlyTestHash}})
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(transplanted)}}
	hashes, diags := readWriteOnlySecretHashes(t.Context(), private)
	require.Empty(t, diags)
	matched, diags := hashes.matches(t.Context(), nested, types.StringValue(strings.Repeat("a", 64)))
	require.Empty(t, diags)
	assert.False(t, matched, "moving a valid verifier must not establish equality at another path")

	for _, emptyPath := range []path.Path{path.Empty(), path.Root("")} {
		require.True(t, hashes.set(t.Context(), emptyPath, types.StringValue("secret")).HasError())
	}

	assert.Equal(t, []byte(transplanted), hashes.prepared)
	assert.Zero(t, private.writes)
}

func TestWriteOnlySecretRejectsInvalidPrivateData(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	cases := map[string][]byte{
		"null":                   []byte(testJSON(nil)),
		"object":                 []byte(testJSON(map[string]any{})),
		"truncated":              []byte(`[`),
		"trailing":               []byte(writeOnlyTestPrivate + `[]`),
		"unknown field":          []byte(testJSON([]map[string]any{{"path": "value_wo", "hash": writeOnlyTestHash, "unexpected": 42}})),
		"empty path":             []byte(testJSON([]writeOnlySecretHashRecord{{Path: "", Hash: writeOnlyTestHash}})),
		"duplicate path":         []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}, {Path: "value_wo", Hash: writeOnlyTestHash}})),
		"unsupported algorithm":  []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "argon2id", "argon2i", 1)}})),
		"unsupported parameters": []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "m=65536", "m=4294967295", 1)}})),
		"invalid key":            []byte(testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "8fzTBn/", "!!!!!!!", 1)}})),
		"excessive size":         bytes.Repeat([]byte(" "), writeOnlySecretHashesMaxLength+1),
		"invalid UTF8":           append([]byte(writeOnlyTestPrivate), 0xff),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
			_, diags := readWriteOnlySecretHashes(t.Context(), private)
			require.True(t, diags.HasError())
			assert.Equal(t, "Invalid write-only secret private state", diags.Errors()[0].Summary())
			assert.NotContains(t, diags.Errors()[0].Detail(), writeOnlyTestHash)
			assert.Zero(t, private.writes)
		})
	}
}

func TestWriteOnlySecretMissingIsNotCorrupt(t *testing.T) {
	t.Parallel()

	for _, data := range [][]byte{nil, []byte(testJSON([]any{}))} {
		private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
		hashes, diags := readWriteOnlySecretHashes(t.Context(), private)
		require.Empty(t, diags)
		matched, diags := hashes.matches(t.Context(), path.Root("value_wo"), types.StringValue(strings.Repeat("a", 64)))
		require.Empty(t, diags)
		assert.False(t, matched)
	}
}

func TestSigningSecretComparisonUsesOneBaseline(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	hashes, diags := readWriteOnlySecretHashes(t.Context(), private)
	require.Empty(t, diags)
	matched, diags := signingSecretValueMatches(t.Context(), &hashes, types.StringValue(strings.Repeat("b", 64)), types.StringValue(strings.Repeat("a", 64)))
	require.Empty(t, diags)
	assert.False(t, matched, "obsolete verifier A must not override known ordinary B")
	matched, diags = signingSecretValueMatches(t.Context(), &hashes, types.StringValue(strings.Repeat("b", 64)), types.StringValue(strings.Repeat("b", 64)))
	require.Empty(t, diags)
	assert.True(t, matched)
}

func TestWriteOnlySecretCancellationDoesNotPrepareOrPublish(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	hashes, diags := readWriteOnlySecretHashes(t.Context(), private)
	require.Empty(t, diags)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.True(t, hashes.set(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("b", 64))).HasError())
	assert.JSONEq(t, writeOnlyTestPrivate, string(hashes.prepared))
	assert.Zero(t, private.writes)
}

func TestSigningSecretPublicationFailurePreservesAcknowledgedState(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})

	for _, test := range []struct {
		name              string
		failPrivateWrite  bool
		invalidModel      bool
		wantPrivateWrites int
	}{
		{name: "private write", failPrivateWrite: true, wantPrivateWrites: 1},
		{name: "public state encoding", invalidModel: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			prior := appSigningSecretTestModel(strings.Repeat("a", 64))
			state := tfsdk.State{Schema: AppSigningSecretResourceSchema(ctx)}
			identity := appSigningSecretTestIdentity(ctx)
			require.Empty(t, setResourceIdentityAndState(ctx, identity, &state, appSigningSecretIdentityAttributeNames(), &prior))
			priorState, priorIdentity := state.Raw.Copy(), identity.Raw.Copy()
			private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}, fail: test.failPrivateWrite}
			hashes, diags := readWriteOnlySecretHashes(ctx, private)
			require.Empty(t, diags)
			require.Empty(t, hashes.set(ctx, path.Root("value_wo"), types.StringValue(strings.Repeat("b", 64))))
			changed := appSigningSecretTestModel(strings.Repeat("b", 64))
			changed.AppDefinitionID = types.StringValue("different-app")

			var model any = &changed
			if test.invalidModel {
				model = "cannot encode a string as resource state"
			}

			diags = publishSigningSecretState(ctx, private, &hashes, identity, &state, appSigningSecretIdentityAttributeNames(), model)
			require.True(t, diags.HasError())
			assert.True(t, priorState.Equal(state.Raw))
			assert.True(t, priorIdentity.Equal(identity.Raw))
			assert.JSONEq(t, writeOnlyTestPrivate, string(private.data[writeOnlySecretHashesPrivateKey]))

			assert.Equal(t, test.wantPrivateWrites, private.writes)

			if test.failPrivateWrite {
				assert.Equal(t, "Private write failed", diags.Errors()[0].Summary())
			}
		})
	}
}

func TestWriteOnlySecretRepeatedPublication(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})
	ctx := t.Context()
	nested := path.Root("headers").AtMapKey("retired-header")
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)
	require.Empty(t, hashes.set(ctx, nested, types.StringValue("second-secret")))

	private.fail = true
	require.True(t, hashes.publish(ctx, private).HasError())
	assert.JSONEq(t, writeOnlyTestPrivate, string(hashes.original))

	private.fail = false
	require.Empty(t, hashes.publish(ctx, private))
	require.Empty(t, hashes.remove(nested))
	require.Empty(t, hashes.publish(ctx, private))
	assert.JSONEq(t, writeOnlyTestPrivate, string(private.data[writeOnlySecretHashesPrivateKey]))
	assert.Equal(t, 3, private.writes)
	require.Empty(t, hashes.publish(ctx, private))
	assert.Equal(t, 3, private.writes, "unchanged publication is skipped")
}

func TestWriteOnlySecretFailedPreparationPreservesStore(t *testing.T) {
	t.Parallel()

	writeOnlyTestPrivate := testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}})
	ctx := t.Context()
	oversized := path.Root(strings.Repeat("z", writeOnlySecretHashesMaxLength))
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(writeOnlyTestPrivate)}}
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)
	require.True(t, hashes.set(ctx, oversized, types.StringValue("candidate")).HasError())
	matched, diags := hashes.matches(ctx, oversized, types.StringValue("candidate"))
	require.Empty(t, diags)
	assert.False(t, matched, "failed candidate must not become a comparison baseline")
	assert.Equal(t, []writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}}, hashes.records)
	assert.JSONEq(t, writeOnlyTestPrivate, string(hashes.prepared))
	assert.JSONEq(t, writeOnlyTestPrivate, string(hashes.original))
	require.Empty(t, hashes.publish(ctx, private))
	assert.Zero(t, private.writes)
}

// Captured from writeWriteOnlySecretHashes at PR #542 commit
// 1d50a229f30d8c85abf163e2c68e1809a3bcad3c, using only synthetic inputs.
func TestWriteOnlySecretPR542CompatibilityAndDynamicPathRemoval(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data, err := os.ReadFile("testdata/write_only_secret_pr542.json")
	require.NoError(t, err)

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
	hashes, diags := readWriteOnlySecretHashes(ctx, private)
	require.Empty(t, diags)

	for _, candidate := range []struct {
		path  path.Path
		value string
	}{
		{path.Root("value_wo"), strings.Repeat("a", 64)},
		{path.Root("header_values_wo").AtMapKey("retired-header"), "pr542-header-secret"},
	} {
		matched, matchDiags := hashes.matches(ctx, candidate.path, types.StringValue(candidate.value))
		require.Empty(t, matchDiags)
		assert.True(t, matched)
	}

	require.Empty(t, hashes.publish(ctx, private))
	assert.Zero(t, private.writes)
	assert.Equal(t, data, private.data[writeOnlySecretHashesPrivateKey])

	var original []writeOnlySecretHashRecord
	require.NoError(t, json.Unmarshal(data, &original))
	// The retired key need not be declared by the current configuration to load
	// its historical record and remove it while retaining the other verifier.
	require.Empty(t, hashes.remove(path.Root("header_values_wo").AtMapKey("retired-header")))
	require.Empty(t, hashes.publish(ctx, private))

	var retained []writeOnlySecretHashRecord
	require.NoError(t, json.Unmarshal(private.data[writeOnlySecretHashesPrivateKey], &retained))
	require.Len(t, retained, 1)
	assert.Equal(t, original[0], retained[0])
}

func TestSigningSecretRejectsOtherVerifierPaths(t *testing.T) {
	t.Parallel()

	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(testJSON([]writeOnlySecretHashRecord{{Path: "other", Hash: writeOnlyTestHash}}))}}
	_, diags := readSigningSecretHashes(t.Context(), private)
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "not supported by this signing-secret resource")
	assert.NotContains(t, diags.Errors()[0].Detail(), writeOnlyTestHash)
	assert.Zero(t, private.writes)
}

func TestWriteOnlySecretDiagnosticsDescribeSafeReasons(t *testing.T) {
	t.Parallel()

	cases := []struct{ data, reason string }{
		{testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "m=65536", "m=42", 1)}}), "unsupported algorithm, version, or parameter encoding"},
		{testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: writeOnlyTestHash}, {Path: "value_wo", Hash: writeOnlyTestHash}}), "duplicate attribute path"},
		{testJSON([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: strings.Replace(writeOnlyTestHash, "8fzTBn/", "!!!!!!!", 1)}}), "invalid salt or key encoding or length"},
		{testJSON(nil), "non-null JSON array"},
	}
	for _, testCase := range cases {
		private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(testCase.data)}}
		_, diags := readWriteOnlySecretHashes(t.Context(), private)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), testCase.reason)

		if strings.HasPrefix(testCase.data, "[") {
			assert.NotContains(t, diags.Errors()[0].Detail(), testCase.data)
		}

		assert.NotContains(t, diags.Errors()[0].Detail(), writeOnlyTestHash)
	}
}
