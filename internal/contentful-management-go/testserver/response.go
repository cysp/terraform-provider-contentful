package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteContentfulManagementResponse(responseWriter http.ResponseWriter, statusCode int, body json.Marshaler) error {
	responseWriter.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
	responseWriter.WriteHeader(statusCode)

	bodyBytes, err := body.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	_, err = responseWriter.Write(bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	return nil
}
