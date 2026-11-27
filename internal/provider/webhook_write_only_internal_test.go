package provider

import (
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// On import, the returned header descriptors are the only redaction evidence.
// Both orderings must reject a readable case alias before map projection.
func TestWebhookRedactedResponseDuplicates(t *testing.T) {
	t.Parallel()

	hidden := cm.WebhookDefinitionHeader{Key: "X-Secret", Secret: cm.NewOptBool(true)}

	readable := cm.WebhookDefinitionHeader{Key: "x-secret", Secret: cm.NewOptBool(false), Value: cm.NewOptString("unexpected-readable-value")}
	for _, headers := range []cm.WebhookDefinitionHeaders{{hidden, readable}, {readable, hidden}} {
		response := cm.WebhookDefinition{Headers: headers}
		sanitized, diags := (webhookRedaction{}).redact(response, WebhookModel{})
		require.True(t, diags.HasError())

		for _, header := range sanitized.Headers {
			require.False(t, header.Value.IsSet())
		}

		require.Contains(t, response.Headers, readable)
	}
}

// Node OpenSSL vectors use the literal credential path, NUL, "synthetic vector"
// and salt bytes 0..15. Swapping two stored hashes must not transfer equality.
func TestWebhookReadsPathBoundVerifiers(t *testing.T) {
	t.Parallel()

	const (
		prefix   = "$argon2id$v=19$m=65536,t=3,p=4$AAECAwQFBgcICQoLDA0ODw$"
		header   = prefix + "Kk5PY/8bzWzfg73Qv0TiL4RGUQsXZlmCf4Y91jt1sok"
		password = prefix + "E+dcEl0zkkeyMPL1tz+y5QT4CbHYUdyJ7++Bq792RQk"
	)

	data := []byte(`{"basic":{"username":"user","hash":"` + password + `"},"headers":{"X-Token.v1":"` + header + `"}}`)
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
	mask := webhookRedaction{password: true, headers: map[string]bool{"X-Token.v1": true}}
	user := types.StringValue("user")
	store, diags := readWebhookVerifierRecords(t.Context(), private, mask, user)
	require.Empty(t, diags)

	value := types.StringValue("synthetic vector")
	match, diags := store.Headers["X-Token.v1"].matches(t.Context(), webhookHeaderVerifierPath("X-Token.v1"), value)
	require.Empty(t, diags)
	require.True(t, match)
	match, diags = store.Basic.Hash.matches(t.Context(), rootWriteOnlySecretPath(webhookPasswordWO), value)
	require.Empty(t, diags)
	require.True(t, match)
	match, diags = store.Basic.Hash.matches(t.Context(), webhookHeaderVerifierPath("X-Token.v1"), value)
	require.Empty(t, diags)
	require.False(t, match)

	change, diags := store.encode()
	require.Empty(t, diags)
	require.Empty(t, change.publish(t.Context(), private))
	require.Zero(t, private.writes)

	// Even if public state already names a different user, this hash cannot
	// acknowledge a password for that user. The header continues untouched.
	store, diags = readWebhookVerifierRecords(t.Context(), private, mask, types.StringValue("other"))
	require.Empty(t, diags)
	require.Nil(t, store.Basic)
	change, diags = store.encode()
	require.Empty(t, diags)
	require.Empty(t, change.publish(t.Context(), private))
	require.JSONEq(t, `{"headers":{"X-Token.v1":"`+header+`"}}`, string(private.data[writeOnlySecretHashesPrivateKey]))
}

func TestWebhookVerifierEnvelope(t *testing.T) {
	t.Parallel()

	for name, data := range map[string]string{
		"null":               `null`,
		"unsupported array":  `[{"path":"http_basic_password_wo","hash":"invalid"}]`,
		"unknown member":     `{"unexpected":[]}`,
		"wrong member case":  `{"Headers":{}}`,
		"duplicate group":    `{"headers":{},"headers":{}}`,
		"duplicate user":     `{"basic":{"username":"a","username":"b","hash":"bad"}}`,
		"duplicate header":   `{"headers":{"X":"a","\u0058":"b"}}`,
		"trailing value":     `{} {}`,
		"basic null":         `{"basic":null}`,
		"username absent":    `{"basic":{"hash":"bad"}}`,
		"username null":      `{"basic":{"username":null,"hash":"bad"}}`,
		"hash absent":        `{"basic":{"username":"a"}}`,
		"headers null":       `{"headers":null}`,
		"header hash null":   `{"headers":{"X":null}}`,
		"header hash number": `{"headers":{"X":1}}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: []byte(data)}}
			_, diags := readWebhookVerifierRecords(t.Context(), private, webhookRedaction{password: true}, types.StringValue("a"))
			require.True(t, diags.HasError())
			require.Zero(t, private.writes)
		})
	}

	// Header names can contain punctuation and differ only in case. Empty
	// usernames are known strings and must not be mistaken for absent fields.
	input := `{"basic":{"username":"","hash":"malformed"},"headers":{"x.y":"one","X.Y":"two","x\"\\":"three"}}`
	v, diags := decodeWebhookVerifiers([]byte(input))
	require.Empty(t, diags)
	require.Equal(t, &webhookBasicVerifier{Username: "", Hash: "malformed"}, v.Basic)
	require.Equal(t, map[string]secretVerifier{"x.y": "one", "X.Y": "two", `x"\`: "three"}, v.Headers)
}

// Only the final collection must fit when renaming a header in nearly full private data.
func TestWebhookVerifierEncodingSizeLimit(t *testing.T) {
	t.Parallel()

	prefix := strings.Repeat("x", writeOnlySecretHashesMaxLength-200)
	oldName, newName := prefix+"A", prefix+"B"
	v := webhookVerifiers{Headers: map[string]secretVerifier{oldName: writeOnlyTestHash}, changed: true}
	_, diags := v.encode()
	require.Empty(t, diags)

	v.Headers[newName] = writeOnlyTestHash
	_, diags = v.encode()
	require.True(t, diags.HasError())
	delete(v.Headers, oldName)
	change, diags := v.encode()
	require.Empty(t, diags)
	require.JSONEq(t, testJSON(map[string]map[string]string{"headers": {newName: writeOnlyTestHash}}), string(change.data))
}
