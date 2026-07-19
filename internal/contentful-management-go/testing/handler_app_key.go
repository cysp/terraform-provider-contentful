package cmtesting

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const (
	maximumAppKeysPerAppDefinition = 3
	defaultAppKeyCollectionLimit   = 100
	maximumAppKeyCollectionLimit   = 1000
	minimumAppKeyX5CEncodedLength  = 736
	maximumAppKeyX5CEncodedLength  = 1416
	minimumAppKeyFingerprintLength = 42
	maximumAppKeyFingerprintLength = 45
	appKeyX5CEncodedPattern        = `^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`
)

//nolint:ireturn
func (ts *Handler) GetAppKeys(_ context.Context, params cm.GetAppKeysParams) (cm.GetAppKeysRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	appKeys := ts.appKeys.List(params.OrganizationID, params.AppDefinitionID)
	skip := int(params.Skip.Or(0))
	limit := int(params.Limit.Or(defaultAppKeyCollectionLimit))

	if skip < 0 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid skip parameter: should be a nonnegative integer"), nil), nil
	}

	if limit < 1 || limit > maximumAppKeyCollectionLimit {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid limit parameter: should be a positive integer lower or equal 1000"), nil), nil
	}

	start := min(skip, len(appKeys))
	end := min(start+limit, len(appKeys))

	return &cm.AppKeyCollection{
		Sys: cm.AppKeyCollectionSys{
			Type: cm.AppKeyCollectionSysTypeArray,
		},
		Total: len(appKeys),
		Skip:  skip,
		Limit: limit,
		Items: appKeys[start:end],
	}, nil
}

//nolint:ireturn
func (ts *Handler) CreateAppKey(_ context.Context, req *cm.AppKeyRequestData, params cm.CreateAppKeyParams) (cm.CreateAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	validationDetails, err := appKeyRequestValidationDetails(*req)
	if err != nil {
		return nil, fmt.Errorf("validate app key request: %w", err)
	}

	if validationDetails != nil {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), validationDetails), nil
	}

	appKey, err := NewAppKeyFromRequest(params.OrganizationID, params.AppDefinitionID, *req)
	if err != nil {
		return nil, err
	}

	if ts.appKeys.Contains(appKey.Sys.ID) {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("The key is already in use"), nil), nil
	}

	if len(ts.appKeys.List(params.OrganizationID, params.AppDefinitionID)) >= maximumAppKeysPerAppDefinition {
		return NewContentfulManagementErrorStatusCodeAccessDenied(new("Forbidden"), []byte(`{"reasons":"Usage exceeded."}`)), nil
	}

	storedAppKey := appKey
	storedAppKey.Generated.Reset()

	ts.appKeys.Set(params.OrganizationID, params.AppDefinitionID, &storedAppKey)

	return &appKey, nil
}

type appKeyValidationError map[string]any

func appKeyRequestValidationDetails(request cm.AppKeyRequestData) ([]byte, error) {
	jwkValue, jwkConfigured, err := decodeAppKeyRequestValue(request.Jwk)
	if err != nil {
		return nil, fmt.Errorf("decode jwk value: %w", err)
	}

	generateValue, generateConfigured, err := decodeAppKeyRequestValue(request.Generate)
	if err != nil {
		return nil, fmt.Errorf("decode generate value: %w", err)
	}

	var errors []appKeyValidationError

	switch {
	case !jwkConfigured && !generateConfigured:
		errors = []appKeyValidationError{
			{"name": "required", "details": `The property "jwk" is required here`, "path": []any{"jwk"}},
			{"name": "required", "details": `The property "generate" is required here`, "path": []any{"generate"}},
			{"name": "unless", "details": `The property "jwk" or "generate" are required here`, "path": []any{}},
		}
	case jwkConfigured && generateConfigured:
		_, jwkIsObject := jwkValue.(map[string]any)
		generate, generateIsTrue := generateValue.(bool)

		if jwkIsObject && generateIsTrue && generate {
			errors = []appKeyValidationError{{"name": "unless", "details": `"jwk" can't be set when "generate" is also set`, "path": []any{}}}
		} else {
			errors = append(errors, appKeyJWKValueValidationErrors(jwkValue)...)
			errors = append(errors, appKeyGenerateValueValidationErrors(generateValue)...)
			errors = append(errors, appKeyValidationError{
				"name":    "unless",
				"details": `The property "jwk" or "generate" are required here`,
				"path":    []any{},
			})
		}
	case jwkConfigured:
		errors = appKeyJWKValueValidationErrors(jwkValue)
	case generateConfigured:
		errors = appKeyGenerateValueValidationErrors(generateValue)
	}

	if len(errors) == 0 {
		return nil, nil
	}

	details, err := json.Marshal(map[string]any{"errors": errors})
	if err != nil {
		return nil, fmt.Errorf("encode validation details: %w", err)
	}

	return details, nil
}

func appKeyJWKValueValidationErrors(value any) []appKeyValidationError {
	jwk, ok := value.(map[string]any)
	if !ok {
		return []appKeyValidationError{appKeyTypeValidationError([]any{"jwk"}, "jwk", "Object", value)}
	}

	return appKeyJWKValidationErrors(jwk)
}

func appKeyGenerateValueValidationErrors(value any) []appKeyValidationError {
	generate, generateIsBoolean := value.(bool)

	var errors []appKeyValidationError

	if !generateIsBoolean {
		errors = append(errors, appKeyTypeValidationError([]any{"generate"}, "generate", "Boolean", value))
	}

	if !generateIsBoolean || !generate {
		errors = append(errors, appKeyValidationError{
			"name":     "in",
			"details":  "Value must be one of expected values",
			"path":     []any{"generate"},
			"value":    value,
			"expected": []any{true},
		})
	}

	return errors
}

func appKeyJWKValidationErrors(jwk map[string]any) []appKeyValidationError {
	var errors []appKeyValidationError

	alg, algConfigured := jwk["alg"]
	errors = append(errors, appKeyJWKEnumValidationErrors(alg, algConfigured, "alg", "RS256")...)

	kty, ktyConfigured := jwk["kty"]
	errors = append(errors, appKeyJWKEnumValidationErrors(kty, ktyConfigured, "kty", "RSA")...)

	use, useConfigured := jwk["use"]
	errors = append(errors, appKeyJWKEnumValidationErrors(use, useConfigured, "use", "sig")...)

	var (
		x5c          string
		publicKeyDER []byte
	)

	requestX5C, x5cConfigured := jwk["x5c"]

	switch requestX5C := requestX5C.(type) {
	case nil:
		if x5cConfigured {
			errors = append(errors, appKeyTypeValidationError([]any{"jwk", "x5c"}, "x5c", "Array", nil))
		} else {
			errors = append(errors, appKeyRequiredValidationError("x5c"))
		}
	case []any:
		switch {
		case len(requestX5C) != 1:
			errors = append(errors, appKeyValidationError{
				"name":    "size",
				"details": "Size must be at least 1 and at most 1",
				"path":    []any{"jwk", "x5c"},
				"value":   requestX5C,
				"min":     1,
				"max":     1,
			})
		default:
			var ok bool

			x5c, ok = requestX5C[0].(string)
			if !ok {
				errors = append(errors, appKeyTypeValidationError(
					[]any{"jwk", "x5c", 0},
					strconv.Itoa(0),
					"String",
					requestX5C[0],
				))

				break
			}

			if len(x5c) < minimumAppKeyX5CEncodedLength || len(x5c) > maximumAppKeyX5CEncodedLength {
				errors = append(errors, appKeyValidationError{
					"name":    "size",
					"details": "Size must be at least 736 and at most 1416",
					"path":    []any{"jwk", "x5c", 0},
					"value":   x5c,
					"min":     minimumAppKeyX5CEncodedLength,
					"max":     maximumAppKeyX5CEncodedLength,
				})
			}

			var valid bool

			publicKeyDER, valid = decodeAppKeyX5CValue(x5c)
			if !valid {
				errors = append(errors, appKeyValidationError{
					"name":    "regexp",
					"details": "Does not match /" + appKeyX5CEncodedPattern + "/",
					"path":    []any{"jwk", "x5c", 0},
					"value":   x5c,
				})
			}
		}
	default:
		errors = append(errors, appKeyTypeValidationError([]any{"jwk", "x5c"}, "x5c", "Array", requestX5C))
	}

	kidValue, kidConfigured := jwk["kid"]
	kid, kidErrors := appKeyJWKFingerprintValidationErrors(kidValue, kidConfigured, "kid")
	errors = append(errors, kidErrors...)

	x5tValue, x5tConfigured := jwk["x5t"]
	x5t, x5tErrors := appKeyJWKFingerprintValidationErrors(x5tValue, x5tConfigured, "x5t")
	errors = append(errors, x5tErrors...)

	if len(errors) != 0 {
		return errors
	}

	fingerprint := cm.AppKeyJWKFingerprint(publicKeyDER)

	if x5t != fingerprint {
		return []appKeyValidationError{{"name": "invalid", "details": "jwk.x5t must be the base64url sha256 fingerprint of the base64 encoded DER public key.", "path": []any{"jwk", "x5t"}}}
	}

	if kid != fingerprint {
		return []appKeyValidationError{{"name": "invalid", "details": "jwk.kid must be the public key sha256 fingerprint. Must match the jwk.x5t.", "path": []any{"jwk", "kid"}}}
	}

	return nil
}

func decodeAppKeyX5CValue(value string) ([]byte, bool) {
	if strings.ContainsAny(value, "\r\n") {
		return nil, false
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, false
	}

	return decoded, true
}

func appKeyJWKEnumValidationErrors(value any, configured bool, property, expected string) []appKeyValidationError {
	if !configured {
		return []appKeyValidationError{appKeyRequiredValidationError(property)}
	}

	stringValue, stringValueValid := value.(string)

	var errors []appKeyValidationError

	if !stringValueValid {
		errors = append(errors, appKeyTypeValidationError([]any{"jwk", property}, property, "String", value))
	}

	if !stringValueValid || stringValue != expected {
		errors = append(errors, appKeyValidationError{
			"name":     "in",
			"details":  "Value must be one of expected values",
			"path":     []any{"jwk", property},
			"value":    value,
			"expected": []any{expected},
		})
	}

	return errors
}

func appKeyJWKFingerprintValidationErrors(value any, configured bool, property string) (string, []appKeyValidationError) {
	if !configured {
		return "", []appKeyValidationError{appKeyRequiredValidationError(property)}
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", []appKeyValidationError{appKeyTypeValidationError([]any{"jwk", property}, property, "String", value)}
	}

	if len(stringValue) < minimumAppKeyFingerprintLength || len(stringValue) > maximumAppKeyFingerprintLength {
		return stringValue, []appKeyValidationError{{
			"name":    "size",
			"details": "Size must be at least 42 and at most 45",
			"path":    []any{"jwk", property},
			"value":   stringValue,
			"min":     minimumAppKeyFingerprintLength,
			"max":     maximumAppKeyFingerprintLength,
		}}
	}

	return stringValue, nil
}

func decodeAppKeyRequestValue(raw []byte) (any, bool, error) {
	if len(raw) == 0 {
		return nil, false, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any

	err := decoder.Decode(&value)
	if err != nil {
		return nil, false, fmt.Errorf("decode JSON: %w", err)
	}

	return value, true, nil
}

func appKeyRequiredValidationError(property string) appKeyValidationError {
	return appKeyValidationError{
		"name":    "required",
		"details": `The property "` + property + `" is required here`,
		"path":    []any{"jwk", property},
	}
}

func appKeyTypeValidationError(path []any, property, expectedType string, value any) appKeyValidationError {
	return appKeyValidationError{
		"name":    "type",
		"details": `The type of "` + property + `" is incorrect, expected type: ` + expectedType,
		"path":    path,
		"type":    expectedType,
		"value":   value,
	}
}

func (ts *Handler) appDefinitionBelongsToOrganization(organizationID, appDefinitionID string) bool {
	appDefinition := ts.appDefinitions[appDefinitionID]

	return appDefinition != nil && appDefinition.Sys.Organization.Sys.ID == organizationID
}

func appDefinitionDoesNotExistError() *cm.ErrorStatusCode {
	return NewContentfulManagementErrorStatusCodeNotFound(new("The resource could not be found."), []byte(`"AppDefinition does not exist."`))
}

//nolint:ireturn
func (ts *Handler) GetAppKey(_ context.Context, params cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	appKey := ts.appKeys.Get(params.OrganizationID, params.AppDefinitionID, params.KeyKid)
	if appKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("The resource could not be found."), []byte(`"AppKey does not exist."`)), nil
	}

	return appKey, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppKey(_ context.Context, params cm.DeleteAppKeyParams) (cm.DeleteAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	appKey := ts.appKeys.Get(params.OrganizationID, params.AppDefinitionID, params.KeyKid)
	if appKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("The resource could not be found."), []byte(`"AppKey does not exist."`)), nil
	}

	ts.appKeys.Delete(params.OrganizationID, params.AppDefinitionID, params.KeyKid)

	return &cm.NoContent{}, nil
}
