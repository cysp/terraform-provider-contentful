package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// jsonObjectValidator checks only the document shape. Callers own member
// semantics and request omission; accepted JSON is never decoded or rewritten.
type jsonObjectValidator struct {
	summary string
	detail  string
}

var _ validator.String = jsonObjectValidator{}

func (v jsonObjectValidator) Description(context.Context) string { return v.detail }
func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validate(req.ConfigValue.ValueString(), req.Path)...)
}

func (v jsonObjectValidator) validate(raw string, valuePath path.Path) diag.Diagnostics {
	if json.Valid([]byte(raw)) && strings.HasPrefix(strings.TrimSpace(raw), "{") {
		return nil
	}

	return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, v.summary, v.detail)}
}
