package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// discoveryIDValidator keeps slash-joined lookup identities unambiguous and
// prevents dot segments from being interpreted as traversal by HTTP routers.
type discoveryIDValidator struct{}

func (discoveryIDValidator) Description(context.Context) string {
	return "Must be nonempty, contain no slash, and not equal a single or double dot."
}

func (v discoveryIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (discoveryIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(req.Path, req.ConfigValue)...)
}

func requireDiscoveryID(attributePath path.Path, value types.String) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attributePath, "Unresolved lookup ID", "The lookup ID must be known and non-null before reading Contentful.")}
	}

	if !validDiscoveryID(value.ValueString()) {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attributePath, "Invalid lookup ID", "The lookup ID must be nonempty, contain no slash, and not equal a single or double dot.")}
	}

	return nil
}

func validDiscoveryID(id string) bool {
	return id != "" && id != "." && id != ".." && !strings.Contains(id, "/")
}

func discoveryResponseIdentityError(detail string) diag.Diagnostics {
	return diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected response identity", detail)}
}

func validateDiscoverySpace(returned, requested string) diag.Diagnostics {
	if returned != requested {
		return discoveryResponseIdentityError("The returned Space ID differs from the requested scope.")
	}

	return nil
}
