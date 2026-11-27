package provider

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type webhookBasicVerifier struct {
	Username string         `json:"username"`
	Hash     secretVerifier `json:"hash"`
}

type webhookVerifiers struct {
	Basic   *webhookBasicVerifier     `json:"basic,omitempty"`
	Headers map[string]secretVerifier `json:"headers,omitempty"`
	changed bool
}

func (v *webhookVerifiers) retain(mask webhookRedaction, username types.String) {
	if v.Basic != nil && (!mask.password || !username.Equal(types.StringValue(v.Basic.Username))) {
		v.Basic = nil
		v.changed = true
	}

	for name := range v.Headers {
		if !mask.headers[name] {
			delete(v.Headers, name)
			v.changed = true
		}
	}
}

func (v *webhookVerifiers) encode() (writeOnlySecretVerifierChange, diag.Diagnostics) {
	if !v.changed || (v.Basic == nil && len(v.Headers) == 0) {
		return writeOnlySecretVerifierChange{changed: v.changed}, nil
	}

	return encodeVerifierChange(v)
}

func readWebhookVerifierRecords(ctx context.Context, private PrivateProviderData, mask webhookRedaction, username types.String) (webhookVerifiers, diag.Diagnostics) {
	data, diags := readVerifierData(ctx, private)
	if diags.HasError() || len(data) == 0 {
		return webhookVerifiers{}, diags
	}

	if !mask.any() {
		return webhookVerifiers{changed: true}, diags
	}

	diags.Append(validateVerifierData(data)...)

	if diags.HasError() {
		return webhookVerifiers{}, diags
	}

	v, decodeDiags := decodeWebhookVerifiers(data)
	diags.Append(decodeDiags...)

	if diags.HasError() {
		return webhookVerifiers{}, diags
	}

	v.retain(mask, username)

	if v.Basic != nil {
		if _, _, reason := parseWriteOnlySecretHash(string(v.Basic.Hash)); reason != "" {
			return webhookVerifiers{}, invalidWriteOnlySecretHashes(reason)
		}
	}

	for _, name := range slices.Sorted(maps.Keys(v.Headers)) {
		if _, _, reason := parseWriteOnlySecretHash(string(v.Headers[name])); reason != "" {
			return webhookVerifiers{}, invalidWriteOnlySecretHashes(reason)
		}
	}

	return v, diags
}

// Read the envelope independently of PHC validation. Refresh can prune one
// credential while carrying an unreadable hash for a continuing sibling.
func decodeWebhookVerifiers(data []byte) (webhookVerifiers, diag.Diagnostics) {
	var fields *struct {
		Basic   jsontext.Value `json:"basic"`
		Headers jsontext.Value `json:"headers"`
	}

	strict := json.RejectUnknownMembers(true)

	err := json.Unmarshal(data, &fields, strict)
	if err != nil || fields == nil {
		return webhookVerifiers{}, invalidWriteOnlySecretHashes("Expected a Webhook verifier object with only basic and headers members and unique names.")
	}

	var v webhookVerifiers

	if fields.Basic != nil {
		var basic struct {
			Username *string         `json:"username"`
			Hash     *secretVerifier `json:"hash"`
		}

		err := json.Unmarshal(fields.Basic, &basic, strict)

		if err != nil || basic.Username == nil || basic.Hash == nil {
			return webhookVerifiers{}, invalidWriteOnlySecretHashes("A Basic verifier requires a string username and hash.")
		}

		v.Basic = &webhookBasicVerifier{Username: *basic.Username, Hash: *basic.Hash}
	}

	if fields.Headers != nil {
		var headers map[string]*secretVerifier

		err := json.Unmarshal(fields.Headers, &headers)
		if err != nil || headers == nil {
			return webhookVerifiers{}, invalidWriteOnlySecretHashes("Expected a Webhook headers object with a string hash for each exact header name.")
		}

		v.Headers = make(map[string]secretVerifier, len(headers))
		for name, hash := range headers {
			if hash == nil {
				return webhookVerifiers{}, invalidWriteOnlySecretHashes("A Webhook header verifier requires a string hash.")
			}

			v.Headers[name] = *hash
		}
	}

	return v, nil
}
