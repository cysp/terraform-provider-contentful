package contentfulmanagementtestserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ReadContentfulManagementRequest(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	return nil
}
