package provider

import (
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

func TestWebhookHeaderVerifierKeyEncoding(t *testing.T) {
	t.Parallel()

	for key, encoded := range map[string]string{
		"X-Field": `headers["X-Field"].value_wo`,
		"x-field": `headers["x-field"].value_wo`,
		`a"b\c`:   `headers["a\"b\\c"].value_wo`,
	} {
		require.Equal(t, encoded, webhookHeaderVerifierPath(key).key)
		decoded, ok := webhookHeaderVerifierKey(encoded)
		require.True(t, ok)
		require.Equal(t, key, decoded)
	}
}

// Node OpenSSL vectors use the literal path, NUL, "synthetic vector" and salt bytes 0..15.
func TestWebhookReadsPathBoundVerifiers(t *testing.T) {
	t.Parallel()

	const (
		prefix = "$argon2id$v=19$m=65536,t=3,p=4$AAECAwQFBgcICQoLDA0ODw$"
		name   = `headers["X-Token.v1"].value_wo`
	)

	header := writeOnlySecretHashRecord{Path: name, Hash: prefix + "Kk5PY/8bzWzfg73Qv0TiL4RGUQsXZlmCf4Y91jt1sok"}
	password := writeOnlySecretHashRecord{Path: "http_basic_password_wo", Hash: prefix + "E+dcEl0zkkeyMPL1tz+y5QT4CbHYUdyJ7++Bq792RQk"}
	data := []byte(testJSON([]writeOnlySecretHashRecord{header, password}))
	private := &writeOnlyTestPrivateData{data: map[string][]byte{writeOnlySecretHashesPrivateKey: data}}
	store, diags := readWriteOnlySecretVerifiers(t.Context(), private, webhookVerifierPath, nil)
	require.Empty(t, diags)

	value := types.StringValue("synthetic vector")
	for _, p := range []writeOnlySecretPath{webhookHeaderVerifierPath("X-Token.v1"), rootWriteOnlySecretPath("http_basic_password_wo")} {
		match, diags := store.matches(t.Context(), p, value)
		require.Empty(t, diags)
		require.True(t, match)
	}

	require.Empty(t, store.publish(t.Context(), private))
	require.Zero(t, private.writes)

	header.Hash = password.Hash
	private.data[writeOnlySecretHashesPrivateKey] = []byte(testJSON([]writeOnlySecretHashRecord{header, password}))
	store, diags = readWriteOnlySecretVerifiers(t.Context(), private, webhookVerifierPath, nil)
	require.Empty(t, diags)
	match, diags := store.matches(t.Context(), webhookHeaderVerifierPath("X-Token.v1"), value)
	require.Empty(t, diags)
	require.False(t, match)

	// A known ordinary password supersedes only its own damaged verifier.
	password.Hash = "obsolete"
	private.data[writeOnlySecretHashesPrivateKey] = []byte(testJSON([]writeOnlySecretHashRecord{header, password}))
	store, diags = readWriteOnlySecretVerifiers(t.Context(), private, webhookVerifierPath, func(p string) bool { return p == password.Path })
	require.Empty(t, diags)
	require.Equal(t, []writeOnlySecretHashRecord{header}, store.records)
	require.Empty(t, store.publish(t.Context(), private))
	require.JSONEq(t, testJSON([]writeOnlySecretHashRecord{header}), string(private.data[writeOnlySecretHashesPrivateKey]))
}
