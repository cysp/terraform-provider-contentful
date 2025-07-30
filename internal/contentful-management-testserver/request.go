package contentfulmanagementtestserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ReadContentfulManagementRequest[T any](r *http.Request, v *T) error {
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

func ReadContentfulManagementRequestWithValidation[T any](r *http.Request, v *T, validateFunc func(*T) error) error {
	readErr := ReadContentfulManagementRequest(r, v)
	if readErr != nil {
		return readErr
	}

	validateErr := validateFunc(v)
	if validateErr != nil {
		return validateErr
	}

	return nil
}
