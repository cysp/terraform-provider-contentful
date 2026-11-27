package provider

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json/v2"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/argon2"
	"golang.org/x/sync/semaphore"
)

const (
	signingSecretWO                 = "value_wo"                 // #nosec G101 -- durable attribute name.
	writeOnlySecretHashesPrivateKey = "write_only_secret_hashes" // #nosec G101 -- private-state key, not a credential.
	writeOnlySecretHashesMaxLength  = 1 << 20
	writeOnlySecretHashMaxLength    = 128
	writeOnlySecretSaltLength       = 16
	writeOnlySecretKeyLength        = 32
	writeOnlySecretMemory           = 64 * 1024
	writeOnlySecretTime             = 3
	writeOnlySecretParallelism      = 4
)

// Share the Argon2 memory budget across resources and provider instances in this process.
var writeOnlySecretHashSlot = semaphore.NewWeighted(1) //nolint:gochecknoglobals

// key identifies the attribute within this resource instance and is included
// in the hash input. attributePath is the Framework path for diagnostics.
type writeOnlySecretPath struct {
	key           string
	attributePath path.Path
}

func rootWriteOnlySecretPath(name string) writeOnlySecretPath {
	return writeOnlySecretPath{key: name, attributePath: path.Root(name)}
}

// An empty verifier means there is no acknowledged write-only value.
type secretVerifier string

// Invalid verifier contents can prevent comparison during planning without
// preventing Core from selecting Create or Delete. Private-data access failures
// and cancellation are different errors and must not receive that treatment.
type invalidWriteOnlySecretHashesDiagnostic struct {
	reason string
}

func (d invalidWriteOnlySecretHashesDiagnostic) Severity() diag.Severity {
	return diag.SeverityError
}

func (d invalidWriteOnlySecretHashesDiagnostic) Summary() string {
	return "Invalid write-only secret private state"
}

func (d invalidWriteOnlySecretHashesDiagnostic) Detail() string {
	return d.reason + " Restore valid Terraform state before updating; the provider will not rotate a secret to repair private state. Explicit replacement or destruction does not require this verifier."
}

func (d invalidWriteOnlySecretHashesDiagnostic) Equal(other diag.Diagnostic) bool {
	invalid, ok := other.(invalidWriteOnlySecretHashesDiagnostic)

	return ok && d.reason == invalid.reason
}

func invalidWriteOnlySecretHashes(reason string) diag.Diagnostics {
	return diag.Diagnostics{invalidWriteOnlySecretHashesDiagnostic{reason: reason}}
}

func readVerifierData(ctx context.Context, private PrivateProviderData) ([]byte, diag.Diagnostics) {
	if private == nil {
		return nil, nil
	}

	return private.GetKey(ctx, writeOnlySecretHashesPrivateKey)
}

func validateVerifierData(data []byte) diag.Diagnostics {
	if len(data) > writeOnlySecretHashesMaxLength {
		return invalidWriteOnlySecretHashes("Verifier data exceeds the size limit.")
	}

	if !utf8.Valid(data) {
		return invalidWriteOnlySecretHashes("Verifier data is not valid UTF-8.")
	}

	return nil
}

func writeOnlySecretHashPrefix() string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", argon2.Version, writeOnlySecretMemory, writeOnlySecretTime, writeOnlySecretParallelism)
}

// The reason is static text: private data must never appear in diagnostics.
func parseWriteOnlySecretHash(encoded string) ([]byte, []byte, string) {
	if len(encoded) > writeOnlySecretHashMaxLength {
		return nil, nil, "A verifier hash exceeds the size limit."
	}

	body, ok := strings.CutPrefix(encoded, writeOnlySecretHashPrefix())
	if !ok {
		return nil, nil, "A verifier uses an unsupported algorithm, version, or parameter encoding."
	}

	saltText, keyText, ok := strings.Cut(body, "$")
	if !ok {
		return nil, nil, "A verifier has an invalid salt and key encoding."
	}

	salt, saltErr := base64.RawStdEncoding.Strict().DecodeString(saltText)

	key, keyErr := base64.RawStdEncoding.Strict().DecodeString(keyText)
	if saltErr != nil || keyErr != nil || len(salt) != writeOnlySecretSaltLength || len(key) != writeOnlySecretKeyLength {
		return nil, nil, "A verifier has an invalid salt or key encoding or length."
	}

	return salt, key, ""
}

func writeOnlySecretKey(ctx context.Context, credential writeOnlySecretPath, value types.String, salt []byte) ([]byte, diag.Diagnostics) {
	err := writeOnlySecretHashSlot.Acquire(ctx, 1)
	if err != nil {
		return nil, writeOnlySecretContextDiagnostics(ctx)
	}
	defer writeOnlySecretHashSlot.Release(1)

	// Bind the verifier to the credential’s path within the resource.
	preimage := []byte(credential.key + "\x00" + value.ValueString())
	key := argon2.IDKey(preimage, salt, writeOnlySecretTime, writeOnlySecretMemory, writeOnlySecretParallelism, writeOnlySecretKeyLength)

	return key, writeOnlySecretContextDiagnostics(ctx)
}

func writeOnlySecretContextDiagnostics(ctx context.Context) diag.Diagnostics {
	if ctx.Err() != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Write-only secret operation cancelled", ctx.Err().Error())}
	}

	return nil
}

func (v secretVerifier) matches(ctx context.Context, credential writeOnlySecretPath, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credential.attributePath, "Unresolved write-only secret", "The secret must be known and non-null before comparison.")}
	}

	if v == "" {
		return false, writeOnlySecretContextDiagnostics(ctx)
	}

	salt, expected, reason := parseWriteOnlySecretHash(string(v))
	if reason != "" {
		return false, invalidWriteOnlySecretHashes(reason)
	}

	actual, diags := writeOnlySecretKey(ctx, credential, value, salt)
	if diags.HasError() {
		return false, diags
	}

	return subtle.ConstantTimeCompare(actual, expected) == 1, diags
}

// A storage-only migration to write-only still needs a fresh verifier. Keep
// exactly the existing salt when both representations are already write-only.
func prepareSecretVerifier(ctx context.Context, verifier secretVerifier, credential writeOnlySecretPath, prior, writeOnly types.String, matches bool) (secretVerifier, diag.Diagnostics) {
	if writeOnly.IsNull() {
		return "", nil
	}

	if writeOnly.IsUnknown() {
		return "", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(credential.attributePath, "Invalid write-only secret verifier input", "A verifier requires a known non-null secret.")}
	}

	if matches && prior.IsNull() {
		return verifier, nil
	}

	salt := make([]byte, writeOnlySecretSaltLength)
	rand.Read(salt)

	key, diags := writeOnlySecretKey(ctx, credential, writeOnly, salt)
	if diags.HasError() {
		return "", diags
	}

	encoded := fmt.Sprintf("%s%s$%s", writeOnlySecretHashPrefix(), base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))

	return secretVerifier(encoded), nil
}

// A change with nil data removes the private key.
type writeOnlySecretVerifierChange struct {
	changed bool
	data    []byte
}

// Finish encoding and checking the proposed state before sending a mutation.
func encodeVerifierChange(value any) (writeOnlySecretVerifierChange, diag.Diagnostics) {
	data, err := json.Marshal(value, json.Deterministic(true))
	if err != nil {
		return writeOnlySecretVerifierChange{}, invalidWriteOnlySecretHashes("The verifiers could not be encoded.")
	}

	if len(data) > writeOnlySecretHashesMaxLength {
		return writeOnlySecretVerifierChange{}, invalidWriteOnlySecretHashes("Verifier data exceeds the size limit.")
	}

	return writeOnlySecretVerifierChange{changed: true, data: data}, nil
}

func (change writeOnlySecretVerifierChange) publish(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	if !change.changed {
		return nil
	}

	if private == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing private state", "The provider cannot publish the acknowledged secret verifiers.")}
	}

	return private.SetKey(ctx, writeOnlySecretHashesPrivateKey, change.data)
}

// A known ordinary value is the authoritative baseline. An obsolete verifier
// must never override a different known ordinary value.
func matchesPriorValue(ctx context.Context, verifier secretVerifier, credential writeOnlySecretPath, priorValue, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() || priorValue.IsUnknown() {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Unresolved secret comparison", "Secret comparison requires a known non-null candidate and a known or null prior value.")}
	}

	if !priorValue.IsNull() {
		return priorValue.Equal(value), nil
	}

	return verifier.matches(ctx, credential, value)
}
