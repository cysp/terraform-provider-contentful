package cmtesting

import (
	"context"
	"errors"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const (
	maximumAppKeysPerAppDefinition = 3
	defaultAppKeyCollectionLimit   = 100
	maximumAppKeyCollectionLimit   = 1000
	minimumAppKeyX5CEncodedLength  = 736
	maximumAppKeyX5CEncodedLength  = 1416
)

var (
	errAppKeyX5CEncodedLength = errors.New("invalid x5c encoded length")
	errAppKeyX5TFingerprint   = errors.New("invalid x5t fingerprint")
	errAppKeyKIDFingerprint   = errors.New("invalid kid fingerprint")
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

	err := validateAppKeyRequest(*req)
	if err != nil {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil), nil
	}

	appKey := NewAppKeyFromRequest(params.OrganizationID, params.AppDefinitionID, *req)

	if ts.appKeys.Contains(appKey.Sys.ID) {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("The key is already in use"), nil), nil
	}

	if len(ts.appKeys.List(params.OrganizationID, params.AppDefinitionID)) >= maximumAppKeysPerAppDefinition {
		return NewContentfulManagementErrorStatusCodeAccessDenied(new("Forbidden"), []byte(`{"reasons":"Usage exceeded."}`)), nil
	}

	ts.appKeys.Set(params.OrganizationID, params.AppDefinitionID, appKey)

	return &appKey, nil
}

func validateAppKeyRequest(request cm.AppKeyRequestData) error {
	err := request.Validate()
	if err != nil {
		return fmt.Errorf("validate request structure: %w", err)
	}

	x5c := request.Jwk.X5c[0]
	if len(x5c) < minimumAppKeyX5CEncodedLength || len(x5c) > maximumAppKeyX5CEncodedLength {
		return errAppKeyX5CEncodedLength
	}

	fingerprint, err := cm.AppKeyJWKFingerprintFromX5C(x5c)
	if err != nil {
		return fmt.Errorf("validate x5c encoding: %w", err)
	}

	if request.Jwk.X5t != fingerprint {
		return errAppKeyX5TFingerprint
	}

	if request.Jwk.Kid != fingerprint {
		return errAppKeyKIDFingerprint
	}

	return nil
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

	appKey, ok := ts.appKeys.Get(params.OrganizationID, params.AppDefinitionID, params.KeyKid)
	if !ok {
		return NewContentfulManagementErrorStatusCodeNotFound(new("The resource could not be found."), []byte(`"AppKey does not exist."`)), nil
	}

	return &appKey, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppKey(_ context.Context, params cm.DeleteAppKeyParams) (cm.DeleteAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.appDefinitionBelongsToOrganization(params.OrganizationID, params.AppDefinitionID) {
		return appDefinitionDoesNotExistError(), nil
	}

	_, ok := ts.appKeys.Get(params.OrganizationID, params.AppDefinitionID, params.KeyKid)
	if !ok {
		return NewContentfulManagementErrorStatusCodeNotFound(new("The resource could not be found."), []byte(`"AppKey does not exist."`)), nil
	}

	ts.appKeys.Delete(params.OrganizationID, params.AppDefinitionID, params.KeyKid)

	return &cm.NoContent{}, nil
}
