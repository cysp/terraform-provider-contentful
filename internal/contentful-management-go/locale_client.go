package contentfulmanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	GetLocaleOperation    OperationName = "GetLocale"
	PutLocaleOperation    OperationName = "PutLocale"
	DeleteLocaleOperation OperationName = "DeleteLocale"
)

type GetLocaleParams struct {
	SpaceID  string
	LocaleID string
}

type PutLocaleParams struct {
	SpaceID            string
	LocaleID           string
	XContentfulVersion OptInt `json:",omitempty,omitzero"`
}

type DeleteLocaleParams struct {
	SpaceID            string
	LocaleID           string
	XContentfulVersion OptInt `json:",omitempty,omitzero"`
}

type GetLocaleRes interface {
	getLocaleRes()
}

type PutLocaleRes interface {
	putLocaleRes()
}

type DeleteLocaleRes interface {
	deleteLocaleRes()
}

type localeErrorRes interface {
	ErrorStatusCodeResponse
	getLocaleRes()
	putLocaleRes()
	deleteLocaleRes()
}

type Locale struct {
	Sys                  LocaleSys    `json:"sys"`
	Name                 string       `json:"name"`
	Code                 string       `json:"code"`
	FallbackCode         OptNilString `json:"fallbackCode"`
	Default              bool         `json:"default"`
	Optional             bool         `json:"optional"`
	ContentDeliveryAPI   bool         `json:"contentDeliveryApi"`
	ContentManagementAPI bool         `json:"contentManagementApi"`
}

type LocaleStatusCode struct {
	StatusCode int
	Response   Locale
}

func (s *LocaleStatusCode) GetStatusCode() int {
	return s.StatusCode
}

func (s *LocaleStatusCode) GetResponse() Locale {
	return s.Response
}

type LocaleRequest struct {
	Name                 string       `json:"name"`
	Code                 string       `json:"code"`
	FallbackCode         OptNilString `json:"fallbackCode,omitempty,omitzero"`
	Optional             OptBool      `json:"optional,omitempty,omitzero"`
	Default              OptBool      `json:"default,omitempty,omitzero"`
	ContentDeliveryAPI   OptBool      `json:"contentDeliveryApi,omitempty,omitzero"`
	ContentManagementAPI OptBool      `json:"contentManagementApi,omitempty,omitzero"`
}

func (*Locale) getLocaleRes()               {}
func (*ApplicationJSONError) getLocaleRes() {}
func (*ApplicationJSONErrorStatusCode) getLocaleRes() {
}
func (*ApplicationVndContentfulManagementV1JSONError) getLocaleRes() {
}
func (*ApplicationVndContentfulManagementV1JSONErrorStatusCode) getLocaleRes() {
}

func (*LocaleStatusCode) putLocaleRes()               {}
func (*ApplicationJSONError) putLocaleRes()           {}
func (*ApplicationJSONErrorStatusCode) putLocaleRes() {}
func (*ApplicationVndContentfulManagementV1JSONError) putLocaleRes() {
}
func (*ApplicationVndContentfulManagementV1JSONErrorStatusCode) putLocaleRes() {
}

func (*NoContent) deleteLocaleRes()                      {}
func (*ApplicationJSONError) deleteLocaleRes()           {}
func (*ApplicationJSONErrorStatusCode) deleteLocaleRes() {}
func (*ApplicationVndContentfulManagementV1JSONError) deleteLocaleRes() {
}
func (*ApplicationVndContentfulManagementV1JSONErrorStatusCode) deleteLocaleRes() {
}

func (c *Client) GetLocale(ctx context.Context, params GetLocaleParams, options ...RequestOption) (GetLocaleRes, error) {
	response, err := c.sendLocaleRequest(ctx, http.MethodGet, params.SpaceID, params.LocaleID, nil, OptInt{}, GetLocaleOperation, options...)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		locale := Locale{}
		if err := decodeLocaleResponseBody(response, &locale); err != nil {
			return nil, err
		}
		return &locale, nil
	default:
		return decodeLocaleErrorResponse(response)
	}
}

func (c *Client) PutLocale(ctx context.Context, request *LocaleRequest, params PutLocaleParams, options ...RequestOption) (PutLocaleRes, error) {
	response, err := c.sendLocaleRequest(ctx, http.MethodPut, params.SpaceID, params.LocaleID, request, params.XContentfulVersion, PutLocaleOperation, options...)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		locale := Locale{}
		if err := decodeLocaleResponseBody(response, &locale); err != nil {
			return nil, err
		}

		return &LocaleStatusCode{
			StatusCode: response.StatusCode,
			Response:   locale,
		}, nil
	default:
		return decodeLocaleErrorResponse(response)
	}
}

func (c *Client) DeleteLocale(ctx context.Context, params DeleteLocaleParams, options ...RequestOption) (DeleteLocaleRes, error) {
	response, err := c.sendLocaleRequest(ctx, http.MethodDelete, params.SpaceID, params.LocaleID, nil, params.XContentfulVersion, DeleteLocaleOperation, options...)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusNoContent:
		if err := response.Body.Close(); err != nil {
			return nil, fmt.Errorf("close response body: %w", err)
		}

		return &NoContent{}, nil
	default:
		return decodeLocaleErrorResponse(response)
	}
}

func (c *Client) sendLocaleRequest(
	ctx context.Context,
	method string,
	spaceID string,
	localeID string,
	body any,
	xContentfulVersion OptInt,
	operationName OperationName,
	requestOptions ...RequestOption,
) (*http.Response, error) {
	var reqCfg requestConfig
	reqCfg.setDefaults(c.baseClient)
	for _, o := range requestOptions {
		o(&reqCfg)
	}

	u := c.serverURL
	if override := reqCfg.ServerURL; override != nil {
		u = override
	}

	u = cloneURL(u)
	u.Path = strings.TrimRight(u.Path, "/") + "/spaces/" + url.PathEscape(spaceID) + "/locales/" + url.PathEscape(localeID)
	u.RawPath = ""

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.contentful.management.v1+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	}

	if value, ok := xContentfulVersion.Get(); ok {
		req.Header.Set("X-Contentful-Version", strconv.Itoa(value))
	}

	if err := c.securityAccessToken(ctx, operationName, req); err != nil {
		return nil, err
	}

	if err := reqCfg.onRequest(req); err != nil {
		return nil, err
	}

	response, err := reqCfg.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if err := reqCfg.onResponse(response); err != nil {
		_ = response.Body.Close()
		return nil, err
	}

	return response, nil
}

func decodeLocaleResponseBody(response *http.Response, out any) error {
	defer func() {
		_ = response.Body.Close()
	}()

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func decodeLocaleErrorResponse(response *http.Response) (localeErrorRes, error) {
	defer func() {
		_ = response.Body.Close()
	}()

	responseError := Error{}
	decodeErr := json.NewDecoder(response.Body).Decode(&responseError)
	if decodeErr != nil && decodeErr != io.EOF {
		return nil, fmt.Errorf("decode error response body: %w", decodeErr)
	}

	contentType := response.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/vnd.contentful.management.v1+json") {
		return &ApplicationVndContentfulManagementV1JSONErrorStatusCode{
			StatusCode: response.StatusCode,
			Response:   NewErrorApplicationVndContentfulManagementV1JSONError(responseError),
		}, nil
	}

	return &ApplicationJSONErrorStatusCode{
		StatusCode: response.StatusCode,
		Response:   NewErrorApplicationJSONError(responseError),
	}, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return &url.URL{}
	}

	urlCopy := *u
	return &urlCopy
}
