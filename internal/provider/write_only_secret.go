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
	"slices"
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
	writeOnlySecretTime             = 1
	writeOnlySecretParallelism      = 4
)

// Share the Argon2 memory budget across resources and provider instances in this process.
var writeOnlySecretHashSlot = semaphore.NewWeighted(1) //nolint:gochecknoglobals

type writeOnlySecretHashRecord struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// Prepare the verifier before PUT; publish it only with acknowledged state.
type writeOnlySecretVerifiers struct {
	original []byte
	prepared []byte
	records  []writeOnlySecretHashRecord
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

func readSigningSecretVerifier(ctx context.Context, private PrivateProviderData, priorValue types.String) (writeOnlySecretVerifiers, diag.Diagnostics) {
	if !priorValue.IsNull() && !priorValue.IsUnknown() {
		return readSupersededWriteOnlySecretVerifiers(ctx, private)
	}

	return readWriteOnlySecretVerifiers(ctx, private, func(p string) bool { return p == "value_wo" }, nil)
}

// All prior comparison values are known in ordinary state. Stage removal of
// even unreadable verifier contents without concealing a private-data access error.
func readSupersededWriteOnlySecretVerifiers(ctx context.Context, private PrivateProviderData) (writeOnlySecretVerifiers, diag.Diagnostics) {
	if private == nil {
		return writeOnlySecretVerifiers{}, nil
	}

	data, diags := private.GetKey(ctx, writeOnlySecretHashesPrivateKey)

	return writeOnlySecretVerifiers{original: bytes.Clone(data)}, diags
}

func readWriteOnlySecretVerifiers(ctx context.Context, private PrivateProviderData, validPath, obsolete func(string) bool) (writeOnlySecretVerifiers, diag.Diagnostics) {
	var store writeOnlySecretVerifiers
	if private == nil {
		return store, nil
	}

	data, diags := private.GetKey(ctx, writeOnlySecretHashesPrivateKey)
	if diags.HasError() || len(data) == 0 {
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

	seen := make(map[string]bool, len(records))

	retained := records[:0]
	for _, record := range records {
		if !validPath(record.Path) {
			return store, invalidWriteOnlySecretHashes("A verifier path is not supported by this resource.")
		}

		if seen[record.Path] {
			return store, invalidWriteOnlySecretHashes("Duplicate verifier paths are not supported.")
		}

		seen[record.Path] = true
		if obsolete != nil && obsolete(record.Path) {
			continue
		}

		if _, _, reason := parseWriteOnlySecretHash(record.Hash); reason != "" {
			return store, invalidWriteOnlySecretHashes(reason)
		}

		retained = append(retained, record)
	}

	store.records = retained
	store.original = bytes.Clone(data)

	store.prepared = store.original
	if len(retained) < len(records) {
		diags.Append(store.prepare(retained)...)
	}

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

func writeOnlySecretKey(ctx context.Context, attributePath path.Path, value types.String, salt []byte) ([]byte, diag.Diagnostics) {
	err := writeOnlySecretHashSlot.Acquire(ctx, 1)
	if err != nil {
		return nil, writeOnlySecretContextDiagnostics(ctx)
	}
	defer writeOnlySecretHashSlot.Release(1)

	// Bind a verifier to its canonical Terraform attribute path.
	preimage := []byte(attributePath.String() + "\x00" + value.ValueString())
	key := argon2.IDKey(preimage, salt, writeOnlySecretTime, writeOnlySecretMemory, writeOnlySecretParallelism, writeOnlySecretKeyLength)

	return key, writeOnlySecretContextDiagnostics(ctx)
}

func writeOnlySecretContextDiagnostics(ctx context.Context) diag.Diagnostics {
	if ctx.Err() != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Write-only secret operation cancelled", ctx.Err().Error())}
	}

	return nil
}

func (s *writeOnlySecretVerifiers) paths() []string {
	paths := make([]string, 0, len(s.records))
	for _, record := range s.records {
		paths = append(paths, record.Path)
	}

	return paths
}

func (s *writeOnlySecretVerifiers) index(attributePath path.Path) int {
	name := attributePath.String()

	return slices.IndexFunc(s.records, func(record writeOnlySecretHashRecord) bool { return record.Path == name })
}

func (s *writeOnlySecretVerifiers) has(attributePath path.Path) bool {
	return s.index(attributePath) >= 0
}

func (s *writeOnlySecretVerifiers) matches(ctx context.Context, attributePath path.Path, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attributePath, "Unresolved write-only secret", "The secret must be known and non-null before comparison.")}
	}

	index := s.index(attributePath)
	if index < 0 {
		return false, writeOnlySecretContextDiagnostics(ctx)
	}

	salt, expected, _ := parseWriteOnlySecretHash(s.records[index].Hash) // Validated on read or prepared locally.

	actual, diags := writeOnlySecretKey(ctx, attributePath, value, salt)
	if diags.HasError() {
		return false, diags
	}

	return subtle.ConstantTimeCompare(actual, expected) == 1, diags
}

// A storage-only migration to write-only still needs a fresh verifier. Keep
// exactly the existing salt when both representations are already write-only.
func (s *writeOnlySecretVerifiers) prepareValue(ctx context.Context, attributePath path.Path, prior, writeOnly types.String, matches bool) diag.Diagnostics {
	if writeOnly.IsNull() {
		return s.remove(attributePath)
	}

	if matches && prior.IsNull() {
		return nil
	}

	return s.set(ctx, attributePath, writeOnly)
}

func (s *writeOnlySecretVerifiers) set(ctx context.Context, attributePath path.Path, value types.String) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attributePath, "Invalid write-only secret verifier input", "A verifier requires a known non-null secret.")}
	}

	salt := make([]byte, writeOnlySecretSaltLength)
	rand.Read(salt)

	key, diags := writeOnlySecretKey(ctx, attributePath, value, salt)
	if diags.HasError() {
		return diags
	}

	encoded := fmt.Sprintf("%s%s$%s", writeOnlySecretHashPrefix(), base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
	records := slices.Clone(s.records)

	record := writeOnlySecretHashRecord{Path: attributePath.String(), Hash: encoded}
	if i := s.index(attributePath); i >= 0 {
		records[i] = record
	} else {
		records = append(records, record)
	}

	return s.prepare(records)
}

func (s *writeOnlySecretVerifiers) remove(attributePath path.Path) diag.Diagnostics {
	index := s.index(attributePath)
	if index < 0 {
		return nil
	}

	return s.prepare(slices.Delete(slices.Clone(s.records), index, index+1))
}

// Finish encoding and enforcing the size bound before making a remote mutation.
func (s *writeOnlySecretVerifiers) prepare(records []writeOnlySecretHashRecord) diag.Diagnostics {
	var data []byte

	if len(records) > 0 {
		slices.SortFunc(records, func(a, b writeOnlySecretHashRecord) int { return strings.Compare(a.Path, b.Path) })

		var err error

		data, err = json.Marshal(records)
		if err != nil {
			return invalidWriteOnlySecretHashes("The verifiers could not be encoded.")
		}

		if len(data) > writeOnlySecretHashesMaxLength {
			return invalidWriteOnlySecretHashes("Verifier data exceeds the size limit.")
		}
	}

	s.records, s.prepared = records, data

	return nil
}

func (s *writeOnlySecretVerifiers) publish(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	if bytes.Equal(s.original, s.prepared) {
		return nil
	}

	if private == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing private state", "The provider cannot publish the acknowledged secret verifiers.")}
	}

	diags := private.SetKey(ctx, writeOnlySecretHashesPrivateKey, s.prepared)
	if !diags.HasError() {
		s.original = s.prepared
	}

	return diags
}
