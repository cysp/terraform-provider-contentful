package provider

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

// Entry lifecycle events describe the operation without serializing content.
// Upstream diagnostic messages remain available separately, including any values
// echoed by Contentful. A decoding error can itself retain a response body.
func entryOperationLogFields(params, response any, err error) map[string]any {
	fields := map[string]any{"params": params}
	if err != nil {
		fields["error_type"] = fmt.Sprintf("%T", err)
	}

	if response, ok := response.(cm.StatusCodeResponse); ok {
		fields["status_code"] = response.GetStatusCode()
	}

	switch response := response.(type) {
	case *cm.EntryStatusCode:
		fields["entry_id"] = response.Response.Sys.ID
		fields["version"] = response.Response.Sys.Version
	case *cm.Entry:
		fields["status_code"] = http.StatusOK
		fields["entry_id"] = response.Sys.ID
		fields["version"] = response.Sys.Version
	case *cm.NoContent:
		fields["status_code"] = http.StatusNoContent
	}

	return fields
}
