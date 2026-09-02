package cmtesting

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
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

	if isMissingWebhookTopics(err) {
		message := "Validation error"
		_ = WriteContentfulManagementErrorResponse(
			w,
			http.StatusUnprocessableEntity,
			"ValidationFailed",
			&message,
			missingWebhookTopicsValidationDetails(),
		)

		return
	}

	ogenerrors.DefaultErrorHandler(ctx, w, r, err)
}

func isMissingWebhookTopics(err error) bool {
	decodeRequestErr, ok := errors.AsType[*ogenerrors.DecodeRequestError](err)
	if !ok || (decodeRequestErr.OperationName() != "CreateWebhookDefinition" && decodeRequestErr.OperationName() != "UpdateWebhookDefinition") {
		return false
	}

	validationErr, ok := errors.AsType[*validate.Error](decodeRequestErr)
	if !ok {
		return false
	}

	for _, field := range validationErr.Fields {
		if field.Name == "topics" && errors.Is(field.Error, validate.ErrFieldRequired) {
			return true
		}
	}

	return false
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
