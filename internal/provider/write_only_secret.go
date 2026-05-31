package provider

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/argon2"
)

const (
	writeOnlySecretHashesPrivateKey = "write_only_secret_hashes" // #nosec G101 -- private state key, not a credential.

	writeOnlySecretArgon2IDMemory      = 64 * 1024
	writeOnlySecretArgon2IDTime        = 1
	writeOnlySecretArgon2IDParallelism = 4
	writeOnlySecretSaltLength          = 16
	writeOnlySecretKeyLength           = 32
	writeOnlySecretParameterPartCount  = 3
)

type privateStateReader interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

type privateStateWriter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

type WriteOnlySecretValue struct {
	Path  path.Path
	Value types.String
}

type WriteOnlySecretValues []WriteOnlySecretValue

func (v *WriteOnlySecretValues) Add(argumentPath path.Path, value types.String) {
	*v = append(*v, WriteOnlySecretValue{
		Path:  argumentPath,
		Value: value,
	})
}

func writeOnlyStringConfigured(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func resolveStringSecret(legacy, writeOnly types.String, legacyPath, writeOnlyPath path.Path, required bool) (types.String, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	legacyConfigured := writeOnlyStringConfigured(legacy)
	writeOnlyConfigured := writeOnlyStringConfigured(writeOnly)

	if legacyConfigured && writeOnlyConfigured {
		diags.AddError(
			"Conflicting secret arguments",
			"Only one of "+legacyPath.String()+" and "+writeOnlyPath.String()+" can be configured.",
		)

		return types.StringNull(), false, diags
	}

	if writeOnlyConfigured {
		return types.StringValue(writeOnly.ValueString()), true, diags
	}

	if legacyConfigured {
		return legacy, false, diags
	}

	if required {
		diags.AddError(
			"Missing secret argument",
			"One of "+legacyPath.String()+" or "+writeOnlyPath.String()+" must be configured.",
		)
	}

	return types.StringNull(), false, diags
}

func writeOnlySecretPreimage(argumentPath path.Path, value types.String) []byte {
	return []byte(argumentPath.String() + "\x00" + value.ValueString())
}

func writeOnlySecretHash(argumentPath path.Path, value types.String) (string, error) {
	salt := make([]byte, writeOnlySecretSaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("generate write-only secret salt: %w", err)
	}

	key := argon2.IDKey(
		writeOnlySecretPreimage(argumentPath, value),
		salt,
		writeOnlySecretArgon2IDTime,
		writeOnlySecretArgon2IDMemory,
		writeOnlySecretArgon2IDParallelism,
		writeOnlySecretKeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		writeOnlySecretArgon2IDMemory,
		writeOnlySecretArgon2IDTime,
		writeOnlySecretArgon2IDParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func writeOnlySecretHashMatches(argumentPath path.Path, value types.String, hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	version, hasVersion := strings.CutPrefix(parts[2], "v=")
	if !hasVersion {
		return false
	}

	parsedVersion, err := strconv.Atoi(version)
	if err != nil || parsedVersion != argon2.Version {
		return false
	}

	parameters, hasMemoryParameter := strings.CutPrefix(parts[3], "m=")
	if !hasMemoryParameter {
		return false
	}

	parameterParts := strings.Split(parameters, ",")
	if len(parameterParts) != writeOnlySecretParameterPartCount {
		return false
	}

	memory, err := strconv.ParseUint(parameterParts[0], 10, 32)
	if err != nil {
		return false
	}

	timePart, hasTimeParameter := strings.CutPrefix(parameterParts[1], "t=")
	if !hasTimeParameter {
		return false
	}

	time, err := strconv.ParseUint(timePart, 10, 32)
	if err != nil {
		return false
	}

	parallelismPart, hasParallelismParameter := strings.CutPrefix(parameterParts[2], "p=")
	if !hasParallelismParameter {
		return false
	}

	parallelism, err := strconv.ParseUint(parallelismPart, 10, 8)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	if len(expectedKey) != writeOnlySecretKeyLength {
		return false
	}

	actualKey := argon2.IDKey(
		writeOnlySecretPreimage(argumentPath, value),
		salt,
		uint32(time),
		uint32(memory),
		uint8(parallelism),
		writeOnlySecretKeyLength,
	)

	return subtle.ConstantTimeCompare(actualKey, expectedKey) == 1
}

type writeOnlySecretHashRecord struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func readWriteOnlySecretHashes(ctx context.Context, private privateStateReader) ([]writeOnlySecretHashRecord, diag.Diagnostics) {
	var diags diag.Diagnostics

	if private == nil {
		return nil, diags
	}

	data, getDiags := private.GetKey(ctx, writeOnlySecretHashesPrivateKey)
	diags.Append(getDiags...)

	if diags.HasError() || len(data) == 0 {
		return nil, diags
	}

	var hashes []writeOnlySecretHashRecord

	err := json.Unmarshal(data, &hashes)
	if err != nil {
		diags.AddError("Invalid write-only secret private state", err.Error())

		return nil, diags
	}

	return hashes, diags
}

func writeWriteOnlySecretHashes(ctx context.Context, private privateStateWriter, values WriteOnlySecretValues) diag.Diagnostics {
	var diags diag.Diagnostics

	if private == nil {
		return diags
	}

	if len(values) == 0 {
		diags.Append(private.SetKey(ctx, writeOnlySecretHashesPrivateKey, nil)...)

		return diags
	}

	hashes := make([]writeOnlySecretHashRecord, 0, len(values))
	for _, value := range values {
		hash, err := writeOnlySecretHash(value.Path, value.Value)
		if err != nil {
			diags.AddError("Failed to create write-only secret hash", err.Error())

			return diags
		}

		hashes = append(hashes, writeOnlySecretHashRecord{
			Path: value.Path.String(),
			Hash: hash,
		})
	}

	data, err := json.Marshal(hashes)
	if err != nil {
		diags.AddError("Invalid write-only secret hashes", err.Error())

		return diags
	}

	diags.Append(private.SetKey(ctx, writeOnlySecretHashesPrivateKey, data)...)

	return diags
}

func writeOnlySecretHashesChanged(ctx context.Context, private privateStateReader, configured WriteOnlySecretValues) (bool, diag.Diagnostics) {
	stored, diags := readWriteOnlySecretHashes(ctx, private)
	if diags.HasError() {
		return false, diags
	}

	if len(stored) != len(configured) {
		return true, diags
	}

	remaining := append([]writeOnlySecretHashRecord(nil), stored...)

	for _, value := range configured {
		matched := false

		for index, record := range remaining {
			if record.Path != value.Path.String() {
				continue
			}

			if !writeOnlySecretHashMatches(value.Path, value.Value, record.Hash) {
				return true, diags
			}

			remaining = append(remaining[:index], remaining[index+1:]...)
			matched = true

			break
		}

		if !matched {
			return true, diags
		}
	}

	return len(remaining) != 0, diags
}

func markWriteOnlySecretChange(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, values WriteOnlySecretValues) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	changed, hashDiags := writeOnlySecretHashesChanged(ctx, req.Private, values)
	resp.Diagnostics.Append(hashDiags...)

	if resp.Diagnostics.HasError() || !changed {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
}
