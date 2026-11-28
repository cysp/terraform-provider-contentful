package cmtesting

import (
	"context"
	"errors"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/ogen-go/ogen/middleware"
)

type issuedAppAccessToken struct {
	installation cm.CreateAppAccessTokenParams
	expiresAt    time.Time
}

type (
	appAccessTokenContextKey struct{}
	appJWTContextKey         struct{}
	authorizedEnvironmentKey struct{}
)

const ValidAccessToken = "CFPAT-12345"

var (
	ErrAccessTokenInvalid = errors.New("AccessTokenInvalid")
	errAppAccessDenied    = errors.New("access denied for app access token")
)

var _ cm.SecurityHandler = (*Handler)(nil)

func (h *Handler) HandleAccessToken(ctx context.Context, _ cm.OperationName, token cm.AccessToken) (context.Context, error) {
	if token.Token == ValidAccessToken {
		return ctx, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	issued, ok := h.appAccessTokens[token.Token]
	if !ok || !time.Now().Before(issued.expiresAt) {
		return nil, ErrAccessTokenInvalid
	}

	return context.WithValue(ctx, appAccessTokenContextKey{}, issued.installation), nil
}

func (*Handler) HandleAppJWT(ctx context.Context, _ cm.OperationName, bearer cm.AppJWT) (context.Context, error) {
	return context.WithValue(ctx, appJWTContextKey{}, bearer.Token), nil
}

func (h *Handler) authorizeAppRequest(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	installation, asApp := req.Context.Value(appAccessTokenContextKey{}).(cm.CreateAppAccessTokenParams)

	if !asApp {
		return next(req)
	}

	switch req.OperationName {
	case cm.GetAppInstallationOperation,
		cm.GetLocalesOperation, cm.GetLocaleOperation, cm.CreateLocaleOperation, cm.PutLocaleOperation, cm.DeleteLocaleOperation,
		cm.GetContentTypesOperation, cm.GetContentTypeOperation, cm.PutContentTypeOperation, cm.DeleteContentTypeOperation,
		cm.ActivateContentTypeOperation, cm.DeactivateContentTypeOperation,
		cm.GetEditorInterfaceOperation, cm.PutEditorInterfaceOperation,
		cm.GetEntriesOperation, cm.GetEntryOperation, cm.CreateEntryOperation, cm.PutEntryOperation, cm.DeleteEntryOperation,
		cm.PublishEntryOperation, cm.UnpublishEntryOperation,
		cm.GetTagOperation, cm.PutTagOperation, cm.DeleteTagOperation:
	default:
		return denyAppOperation(req.OperationName)
	}

	space, _ := req.Params.Path("space_id")
	environment, _ := req.Params.Path("environment_id")
	spaceID, spaceOK := space.(string)
	environmentID, environmentOK := environment.(string)

	if !spaceOK || !environmentOK {
		return middleware.Response{}, errAppAccessDenied
	}

	h.mu.Lock()
	target := h.resolveEnvironmentAlias(spaceID, environmentID)
	h.mu.Unlock()

	app, hasApp := req.Params.Path("app_definition_id")
	if spaceID != installation.SpaceID || target != installation.EnvironmentID || (hasApp && app != installation.AppDefinitionID) {
		return middleware.Response{}, errAppAccessDenied
	}

	req.SetContext(context.WithValue(req.Context, authorizedEnvironmentKey{}, target))

	return next(req)
}

func denyAppOperation(operation cm.OperationName) (middleware.Response, error) {
	switch operation {
	case cm.GetAppDefinitionOperation:
		return middleware.Response{Type: NewContentfulManagementErrorStatusCodeNotFound(nil, nil)}, nil
	case cm.GetSpaceOperation, cm.GetEnvironmentOperation:
		return middleware.Response{}, ErrAccessTokenInvalid
	default:
		return middleware.Response{}, errAppAccessDenied
	}
}
