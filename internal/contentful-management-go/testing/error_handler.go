package cmtesting

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ogen-go/ogen/ogenerrors"
)

func errorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if _, ok := errors.AsType[*ogenerrors.SecurityError](err); ok {
		message := "The access token you sent could not be found or is invalid."
		_ = WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", &message, nil)

		return
	}

	if isMissingTaxonomyDeleteVersion(r, err) {
		const detail = "Invalid input: expected number, received NaN"

		details, detailsErr := taxonomyDeleteVersionValidationDetails("invalid_type", detail)
		if detailsErr == nil {
			message := "Validation error"
			_ = WriteContentfulManagementErrorResponse(w, http.StatusUnprocessableEntity, "ValidationFailed", &message, details)

			return
		}
	}

	ogenerrors.DefaultErrorHandler(ctx, w, r, err)
}

func isMissingTaxonomyDeleteVersion(r *http.Request, err error) bool {
	if r.Method != http.MethodDelete || len(r.Header.Values("X-Contentful-Version")) != 0 {
		return false
	}

	if !strings.Contains(r.URL.Path, "/taxonomy/concepts/") && !strings.Contains(r.URL.Path, "/taxonomy/concept-schemes/") {
		return false
	}

	var decodeParamsErr *ogenerrors.DecodeParamsError

	var decodeParamErr *ogenerrors.DecodeParamError

	return errors.As(err, &decodeParamsErr) && errors.As(err, &decodeParamErr) && decodeParamErr.Name == "X-Contentful-Version"
}
