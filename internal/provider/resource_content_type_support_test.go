package provider_test

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var (
	errUnexpectedContentTypeRequestCount        = errors.New("unexpected content type request count")
	errUnexpectedContentTypeResponseType        = errors.New("unexpected content type response type")
	errUnexpectedContentTypePublicationVersions = errors.New("unexpected content type publication versions")
)

func requireNoConfirmingGetAfterActivationFailure(t *testing.T, handler *contentTypeActivationTestHandler) {
	t.Helper()
	require.Equal(t, 0, handler.eventCount(contentTypeOperationGet), "provider issued a confirming GET after an activation failure")
	handler.activationFailureReturned.Store(false)
}

func contentTypeActivationConfigVariables(contentTypeID string) config.Variables {
	return config.Variables{
		"space_id":             config.StringVariable("space"),
		"environment_id":       config.StringVariable("environment"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}
}

func contentTypeActivationRequestCheck(handler *contentTypeActivationTestHandler, wantPuts, wantActivations int64) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := int64(handler.eventCount(contentTypeOperationPut)); got != wantPuts {
			return fmt.Errorf("%w: got %d content-type PUTs, want %d", errUnexpectedContentTypeRequestCount, got, wantPuts)
		}

		if got := int64(handler.eventCount(contentTypeOperationActivate)); got != wantActivations {
			return fmt.Errorf("%w: got %d activations, want %d", errUnexpectedContentTypeRequestCount, got, wantActivations)
		}

		for _, request := range handler.rawRequestHistory() {
			switch request.Operation {
			case contentTypeOperationGet:
			case contentTypeOperationPut:
				if request.Method != http.MethodPut || len(request.Body) == 0 || request.ContentLength <= 0 {
					return fmt.Errorf("%w: malformed content-type draft request: method=%s body=%d content-length=%d", errUnexpectedContentTypeRequestCount, request.Method, len(request.Body), request.ContentLength)
				}

				if len(request.VersionValues) > 1 {
					return fmt.Errorf("%w: content-type draft request has %d version headers", errUnexpectedContentTypeRequestCount, len(request.VersionValues))
				}
			case contentTypeOperationActivate:
				if request.Method != http.MethodPut || len(request.Body) != 0 || request.ContentLength != 0 {
					return fmt.Errorf("%w: malformed content-type activation request: method=%s body=%d content-length=%d", errUnexpectedContentTypeRequestCount, request.Method, len(request.Body), request.ContentLength)
				}

				if len(request.VersionValues) != 1 || request.VersionValues[0] != strconv.FormatInt(request.Version, 10) {
					return fmt.Errorf("%w: content-type activation version headers=%v parsed-version=%d", errUnexpectedContentTypePublicationVersions, request.VersionValues, request.Version)
				}
			}
		}

		return nil
	}
}

func contentTypeActivationRequestAndVersionsCheck(
	handler *contentTypeActivationTestHandler,
	wantPuts int64,
	wantActivations int64,
	wantActivationVersions []int64,
) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		contentTypeActivationRequestCheck(handler, wantPuts, wantActivations),
		func(_ *terraform.State) error {
			if got := handler.activationVersionHistory(); !slices.Equal(got, wantActivationVersions) {
				return fmt.Errorf(
					"%w: got activation versions %v, want %v",
					errUnexpectedContentTypePublicationVersions,
					got,
					wantActivationVersions,
				)
			}

			wantOperations := make([]contentTypeOperation, 0, wantPuts+wantActivations)
			for range wantPuts {
				wantOperations = append(wantOperations, contentTypeOperationPut)
			}

			for range wantActivations {
				wantOperations = append(wantOperations, contentTypeOperationActivate)
			}

			gotOperations := make([]contentTypeOperation, 0, len(wantOperations))

			for _, event := range handler.eventHistory() {
				if event.Operation == contentTypeOperationPut || event.Operation == contentTypeOperationActivate {
					gotOperations = append(gotOperations, event.Operation)
				}
			}

			if !slices.Equal(gotOperations, wantOperations) {
				return fmt.Errorf(
					"%w: got operation order %v, want %v",
					errUnexpectedContentTypeRequestCount,
					gotOperations,
					wantOperations,
				)
			}

			return nil
		},
	)
}
