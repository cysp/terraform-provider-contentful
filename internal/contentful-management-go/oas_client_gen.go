// Code generated by ogen, DO NOT EDIT.

package client

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/conv"
	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
)

// Invoker invokes operations described by OpenAPI v3 specification.
type Invoker interface {
	// DeleteAppInstallation invokes deleteAppInstallation operation.
	//
	// Uninstall an app.
	//
	// DELETE /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
	DeleteAppInstallation(ctx context.Context, params DeleteAppInstallationParams) (DeleteAppInstallationRes, error)
	// GetAppDefinition invokes getAppDefinition operation.
	//
	// Get one app definition.
	//
	// GET /organizations/{organization_id}/app_definitions/{app_definition_id}
	GetAppDefinition(ctx context.Context, params GetAppDefinitionParams) (GetAppDefinitionRes, error)
	// GetAppInstallation invokes getAppInstallation operation.
	//
	// Get one app installation.
	//
	// GET /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
	GetAppInstallation(ctx context.Context, params GetAppInstallationParams) (GetAppInstallationRes, error)
	// GetAuthenticatedUser invokes getAuthenticatedUser operation.
	//
	// Get the authenticated user.
	//
	// GET /users/me
	GetAuthenticatedUser(ctx context.Context) (GetAuthenticatedUserRes, error)
	// GetOrganization invokes getOrganization operation.
	//
	// Get an organization an admin or owner has access to.
	//
	// GET /organizations/{organization_id}
	GetOrganization(ctx context.Context, params GetOrganizationParams) (GetOrganizationRes, error)
	// PutAppInstallation invokes putAppInstallation operation.
	//
	// Install or update an app.
	//
	// PUT /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
	PutAppInstallation(ctx context.Context, request *PutAppInstallationReq, params PutAppInstallationParams) (PutAppInstallationRes, error)
}

// Client implements OAS client.
type Client struct {
	serverURL *url.URL
	sec       SecuritySource
	baseClient
}

func trimTrailingSlashes(u *url.URL) {
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = strings.TrimRight(u.RawPath, "/")
}

// NewClient initializes new Client defined by OAS.
func NewClient(serverURL string, sec SecuritySource, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	trimTrailingSlashes(u)

	c, err := newClientConfig(opts...).baseClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		serverURL:  u,
		sec:        sec,
		baseClient: c,
	}, nil
}

type serverURLKey struct{}

// WithServerURL sets context key to override server URL.
func WithServerURL(ctx context.Context, u *url.URL) context.Context {
	return context.WithValue(ctx, serverURLKey{}, u)
}

func (c *Client) requestURL(ctx context.Context) *url.URL {
	u, ok := ctx.Value(serverURLKey{}).(*url.URL)
	if !ok {
		return c.serverURL
	}
	return u
}

// DeleteAppInstallation invokes deleteAppInstallation operation.
//
// Uninstall an app.
//
// DELETE /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
func (c *Client) DeleteAppInstallation(ctx context.Context, params DeleteAppInstallationParams) (DeleteAppInstallationRes, error) {
	res, err := c.sendDeleteAppInstallation(ctx, params)
	return res, err
}

func (c *Client) sendDeleteAppInstallation(ctx context.Context, params DeleteAppInstallationParams) (res DeleteAppInstallationRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [6]string
	pathParts[0] = "/spaces/"
	{
		// Encode "space_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "space_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.SpaceID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	pathParts[2] = "/environments/"
	{
		// Encode "environment_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "environment_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.EnvironmentID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[3] = encoded
	}
	pathParts[4] = "/app_installations/"
	{
		// Encode "app_definition_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "app_definition_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.AppDefinitionID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[5] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "DELETE", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "DeleteAppInstallation", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodeDeleteAppInstallationResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// GetAppDefinition invokes getAppDefinition operation.
//
// Get one app definition.
//
// GET /organizations/{organization_id}/app_definitions/{app_definition_id}
func (c *Client) GetAppDefinition(ctx context.Context, params GetAppDefinitionParams) (GetAppDefinitionRes, error) {
	res, err := c.sendGetAppDefinition(ctx, params)
	return res, err
}

func (c *Client) sendGetAppDefinition(ctx context.Context, params GetAppDefinitionParams) (res GetAppDefinitionRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [4]string
	pathParts[0] = "/organizations/"
	{
		// Encode "organization_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "organization_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.OrganizationID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	pathParts[2] = "/app_definitions/"
	{
		// Encode "app_definition_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "app_definition_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.AppDefinitionID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[3] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "GET", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "GetAppDefinition", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodeGetAppDefinitionResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// GetAppInstallation invokes getAppInstallation operation.
//
// Get one app installation.
//
// GET /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
func (c *Client) GetAppInstallation(ctx context.Context, params GetAppInstallationParams) (GetAppInstallationRes, error) {
	res, err := c.sendGetAppInstallation(ctx, params)
	return res, err
}

func (c *Client) sendGetAppInstallation(ctx context.Context, params GetAppInstallationParams) (res GetAppInstallationRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [6]string
	pathParts[0] = "/spaces/"
	{
		// Encode "space_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "space_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.SpaceID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	pathParts[2] = "/environments/"
	{
		// Encode "environment_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "environment_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.EnvironmentID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[3] = encoded
	}
	pathParts[4] = "/app_installations/"
	{
		// Encode "app_definition_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "app_definition_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.AppDefinitionID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[5] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "GET", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "GetAppInstallation", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodeGetAppInstallationResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// GetAuthenticatedUser invokes getAuthenticatedUser operation.
//
// Get the authenticated user.
//
// GET /users/me
func (c *Client) GetAuthenticatedUser(ctx context.Context) (GetAuthenticatedUserRes, error) {
	res, err := c.sendGetAuthenticatedUser(ctx)
	return res, err
}

func (c *Client) sendGetAuthenticatedUser(ctx context.Context) (res GetAuthenticatedUserRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/users/me"
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "GET", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "GetAuthenticatedUser", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodeGetAuthenticatedUserResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// GetOrganization invokes getOrganization operation.
//
// Get an organization an admin or owner has access to.
//
// GET /organizations/{organization_id}
func (c *Client) GetOrganization(ctx context.Context, params GetOrganizationParams) (GetOrganizationRes, error) {
	res, err := c.sendGetOrganization(ctx, params)
	return res, err
}

func (c *Client) sendGetOrganization(ctx context.Context, params GetOrganizationParams) (res GetOrganizationRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [2]string
	pathParts[0] = "/organizations/"
	{
		// Encode "organization_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "organization_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.OrganizationID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "GET", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "GetOrganization", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodeGetOrganizationResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// PutAppInstallation invokes putAppInstallation operation.
//
// Install or update an app.
//
// PUT /spaces/{space_id}/environments/{environment_id}/app_installations/{app_definition_id}
func (c *Client) PutAppInstallation(ctx context.Context, request *PutAppInstallationReq, params PutAppInstallationParams) (PutAppInstallationRes, error) {
	res, err := c.sendPutAppInstallation(ctx, request, params)
	return res, err
}

func (c *Client) sendPutAppInstallation(ctx context.Context, request *PutAppInstallationReq, params PutAppInstallationParams) (res PutAppInstallationRes, err error) {

	u := uri.Clone(c.requestURL(ctx))
	var pathParts [6]string
	pathParts[0] = "/spaces/"
	{
		// Encode "space_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "space_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.SpaceID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	pathParts[2] = "/environments/"
	{
		// Encode "environment_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "environment_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.EnvironmentID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[3] = encoded
	}
	pathParts[4] = "/app_installations/"
	{
		// Encode "app_definition_id" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "app_definition_id",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.StringToString(params.AppDefinitionID))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[5] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	r, err := ht.NewRequest(ctx, "PUT", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodePutAppInstallationRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	{
		type bitset = [1]uint8
		var satisfied bitset
		{

			switch err := c.securityAccessToken(ctx, "PutAppInstallation", r); {
			case err == nil: // if NO error
				satisfied[0] |= 1 << 0
			case errors.Is(err, ogenerrors.ErrSkipClientSecurity):
				// Skip this security.
			default:
				return res, errors.Wrap(err, "security \"AccessToken\"")
			}
		}

		if ok := func() bool {
		nextRequirement:
			for _, requirement := range []bitset{
				{0b00000001},
			} {
				for i, mask := range requirement {
					if satisfied[i]&mask != mask {
						continue nextRequirement
					}
				}
				return true
			}
			return false
		}(); !ok {
			return res, ogenerrors.ErrSecurityRequirementIsNotSatisfied
		}
	}

	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	result, err := decodePutAppInstallationResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}
