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

type writeOnlySecretHashRecord struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// This store only handles bounded path-indexed verifier data. Resources own
// lifecycle decisions and publish prepared data only with acknowledged state.
type writeOnlySecretHashes struct {
	original []byte
	prepared []byte
	records  []writeOnlySecretHashRecord
}

func invalidWriteOnlySecretHashes(reason string) diag.Diagnostics {
	return diag.Diagnostics{diag.NewErrorDiagnostic("Invalid write-only secret private state", reason+" Restore valid Terraform state before retrying; the provider will not rotate a secret to repair private state.")}
}

func readWriteOnlySecretHashes(ctx context.Context, private PrivateProviderData) (writeOnlySecretHashes, diag.Diagnostics) {
	store := writeOnlySecretHashes{}

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

	err := decoder.Decode(&store.records)
	if err != nil || store.records == nil {
		return store, invalidWriteOnlySecretHashes("Expected a non-null JSON array of path and hash records with no additional fields.")
	}

	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return store, invalidWriteOnlySecretHashes("Unexpected data follows the verifier records.")
	}

	seen := make(map[string]bool, len(store.records))
	for _, record := range store.records {
		if record.Path == "" {
			return store, invalidWriteOnlySecretHashes("A verifier record has an empty attribute path.")
		}

		if seen[record.Path] {
			return store, invalidWriteOnlySecretHashes("Verifier records contain a duplicate attribute path.")
		}

		_, _, reason := parseWriteOnlySecretHash(record.Hash)
		if reason != "" {
			return store, invalidWriteOnlySecretHashes(reason)
		}

		seen[record.Path] = true
	}

	store.original = bytes.Clone(data)
	store.prepared = bytes.Clone(data)

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

func writeOnlySecretKey(argumentPath path.Path, value types.String, salt []byte) []byte {
	// Bind a verifier to its canonical Terraform attribute path.
	preimage := []byte(argumentPath.String() + "\x00" + value.ValueString())

	return argon2.IDKey(preimage, salt, writeOnlySecretTime, writeOnlySecretMemory, writeOnlySecretParallelism, writeOnlySecretKeyLength)
}

func writeOnlySecretContextDiagnostics(ctx context.Context) diag.Diagnostics {
	if ctx.Err() != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Write-only secret operation cancelled", ctx.Err().Error())}
	}

	return nil
}

func (s *writeOnlySecretHashes) matches(ctx context.Context, argumentPath path.Path, value types.String) (bool, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return false, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(argumentPath, "Unresolved signing secret", "The secret must be known and non-null before comparison.")}
	}

	diags := writeOnlySecretContextDiagnostics(ctx)
	if diags.HasError() {
		return false, diags
	}

	for _, record := range s.records {
		if record.Path != argumentPath.String() {
			continue
		}

		salt, expected, reason := parseWriteOnlySecretHash(record.Hash)
		if reason != "" {
			return false, invalidWriteOnlySecretHashes(reason)
		}

		actual := writeOnlySecretKey(argumentPath, value, salt)

		diags.Append(writeOnlySecretContextDiagnostics(ctx)...)

		return subtle.ConstantTimeCompare(actual, expected) == 1, diags
	}

	return false, diags
}

// set prepares a replacement with a fresh salt. Callers compare first when
// equality should preserve the existing verifier. A failure leaves the store unchanged.
func (s *writeOnlySecretHashes) set(ctx context.Context, argumentPath path.Path, value types.String) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() || argumentPath.String() == "" {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(argumentPath, "Invalid write-only secret verifier input", "A verifier requires a known non-null value at a non-empty attribute path.")}
	}

	diags := writeOnlySecretContextDiagnostics(ctx)
	if diags.HasError() {
		return diags
	}

	salt := make([]byte, writeOnlySecretSaltLength)

	rand.Read(salt)

	key := writeOnlySecretKey(argumentPath, value, salt)

	diags.Append(writeOnlySecretContextDiagnostics(ctx)...)

	if diags.HasError() {
		return diags
	}

	encoded := fmt.Sprintf("%s%s$%s", writeOnlySecretHashPrefix(), base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
	records := slices.DeleteFunc(slices.Clone(s.records), func(record writeOnlySecretHashRecord) bool { return record.Path == argumentPath.String() })
	records = append(records, writeOnlySecretHashRecord{Path: argumentPath.String(), Hash: encoded})

	return s.prepare(records)
}

func (s *writeOnlySecretHashes) remove(argumentPath path.Path) diag.Diagnostics {
	records := slices.DeleteFunc(slices.Clone(s.records), func(record writeOnlySecretHashRecord) bool { return record.Path == argumentPath.String() })
	if len(records) == len(s.records) {
		return nil
	}

	return s.prepare(records)
}

func (s *writeOnlySecretHashes) prepare(records []writeOnlySecretHashRecord) diag.Diagnostics {
	var data []byte

	if len(records) != 0 {
		slices.SortFunc(records, func(a, b writeOnlySecretHashRecord) int { return strings.Compare(a.Path, b.Path) })

		var err error

		data, err = json.Marshal(records)
		if err != nil {
			return invalidWriteOnlySecretHashes("Verifier records could not be encoded.")
		}

		if len(data) > writeOnlySecretHashesMaxLength {
			return invalidWriteOnlySecretHashes("Verifier data exceeds the size limit.")
		}
	}

	s.records = records
	s.prepared = data

	return nil
}

func (s *writeOnlySecretHashes) publish(ctx context.Context, private PrivateProviderData) diag.Diagnostics {
	if bytes.Equal(s.original, s.prepared) {
		return nil
	}

	if private == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Missing private state", "The provider cannot publish the acknowledged signing secret verifier.")}
	}

	diags := private.SetKey(ctx, writeOnlySecretHashesPrivateKey, s.prepared)
	if !diags.HasError() {
		s.original = bytes.Clone(s.prepared)
	}

	return diags
}
