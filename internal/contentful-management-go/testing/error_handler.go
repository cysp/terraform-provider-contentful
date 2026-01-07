package testing

import (
	"context"
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"
)

func errorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	var securityErr *ogenerrors.SecurityError
	if errors.As(err, &securityErr) {
		message := "The access token you sent could not be found or is invalid."
		_ = WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", &message, nil)

		return
	}

	ogenerrors.DefaultErrorHandler(ctx, w, r, err)
}
