package cmtesting

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteContentfulManagementResponse(w http.ResponseWriter, statusCode int, body any) error {
	w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
	w.WriteHeader(statusCode)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	_, err = w.Write(bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	return nil
}
