package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/argon2"
	"golang.org/x/sync/semaphore"
)

const (
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
var signingSecretHashSlot = semaphore.NewWeighted(1) //nolint:gochecknoglobals

type writeOnlySecretHashRecord struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// Prepare the verifier before PUT; publish it only with acknowledged state.
type signingSecretVerifier struct {
	original []byte
	prepared []byte
	salt     []byte
	key      []byte
}

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

func readSigningSecretVerifier(ctx context.Context, private PrivateProviderData, priorValue types.String) (signingSecretVerifier, diag.Diagnostics) {
	store := signingSecretVerifier{}

	if private == nil {
		return store, nil
	}

	data, diags := private.GetKey(ctx, writeOnlySecretHashesPrivateKey)
	if diags.HasError() || len(data) == 0 {
		return store, diags
	}

	// An ordinary value supplies the comparison. Retain obsolete private data
	// only to stage its removal or replacement at publication.
	if !priorValue.IsNull() && !priorValue.IsUnknown() {
		store.original = bytes.Clone(data)

		return store, diags
	}

	if len(data) > writeOnlySecretHashesMaxLength {
		return store, invalidWriteOnlySecretHashes("Verifier data exceeds the size limit.")
	}

	if !utf8.Valid(data) {
		return store, invalidWriteOnlySecretHashes("Verifier data is not valid UTF-8.")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var records []writeOnlySecretHashRecord

	err := decoder.Decode(&records)
	if err != nil || records == nil {
		return store, invalidWriteOnlySecretHashes("Expected a non-null JSON array of path and hash records with no additional fields.")
	}

	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return store, invalidWriteOnlySecretHashes("Unexpected data follows the verifier records.")
	}

	if len(records) > 1 {
		return store, invalidWriteOnlySecretHashes("Expected at most one signing secret verifier.")
	}

	if len(records) == 1 {
		if records[0].Path != "value_wo" {
			return store, invalidWriteOnlySecretHashes("A verifier path is not supported by this signing-secret resource.")
		}

		salt, key, reason := parseWriteOnlySecretHash(records[0].Hash)
		if reason != "" {
			return store, invalidWriteOnlySecretHashes(reason)
		}

		store.salt, store.key = salt, key
	}

	store.original = bytes.Clone(data)
	store.prepared = store.original

	return store, diags
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

func writeOnlySecretKey(ctx context.Context, value types.String, salt []byte) ([]byte, diag.Diagnostics) {
	err := signingSecretHashSlot.Acquire(ctx, 1)
	if err != nil {
		return nil, writeOnlySecretContextDiagnostics(ctx)
	}
	defer signingSecretHashSlot.Release(1)

	// Bind a verifier to its canonical Terraform attribute path.
	preimage := []byte("value_wo\x00" + value.ValueString())
	key := argon2.IDKey(preimage, salt, writeOnlySecretTime, writeOnlySecretMemory, writeOnlySecretParallelism, writeOnlySecretKeyLength)

	return key, writeOnlySecretContextDiagnostics(ctx)
}

func writeOnlySecretContextDiagnostics(ctx context.Context) diag.Diagnostics {
	if ctx.Err() != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Write-only secret operation cancelled", ctx.Err().Error())}
	}

	return nil
}

func (s *signingSecretVerifier) matches(ctx context.Context, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("value_wo"), "Unresolved signing secret", "The secret must be known and non-null before comparison.")}
	}

	if s.key == nil {
		return false, writeOnlySecretContextDiagnostics(ctx)
	}

	actual, diags := writeOnlySecretKey(ctx, value, s.salt)
	if diags.HasError() {
		return false, diags
	}

	return subtle.ConstantTimeCompare(actual, s.key) == 1, diags
}

// set prepares a replacement with a fresh salt. Callers compare first when
// equality should preserve the existing verifier. A failure leaves the store unchanged.
func (s *signingSecretVerifier) set(ctx context.Context, value types.String) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("value_wo"), "Invalid write-only secret verifier input", "A verifier requires a known non-null signing secret.")}
	}

	salt := make([]byte, writeOnlySecretSaltLength)

	rand.Read(salt)

	key, diags := writeOnlySecretKey(ctx, value, salt)
	if diags.HasError() {
		return diags
	}

	encoded := fmt.Sprintf("%s%s$%s", writeOnlySecretHashPrefix(), base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))

	data, err := json.Marshal([]writeOnlySecretHashRecord{{Path: "value_wo", Hash: encoded}})
	if err != nil {
		return invalidWriteOnlySecretHashes("The verifier could not be encoded.")
	}

	s.salt, s.key, s.prepared = salt, key, data

	return nil
}

func (s *signingSecretVerifier) clear() {
	s.salt, s.key, s.prepared = nil, nil, nil
}

func (s *signingSecretVerifier) publish(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	if bytes.Equal(s.original, s.prepared) {
		return nil
	}

	if private == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing private state", "The provider cannot publish the acknowledged signing secret verifier.")}
	}

	diags := private.SetKey(ctx, writeOnlySecretHashesPrivateKey, s.prepared)
	if !diags.HasError() {
		s.original = s.prepared
	}

	return diags
}
