// Code generated by ogen, DO NOT EDIT.

package client

import (
	"bytes"
	"net/http"

	"github.com/go-faster/jx"

	ht "github.com/ogen-go/ogen/http"
)

func encodeCreateDeliveryApiKeyRequest(
	req *CreateDeliveryApiKeyReq,
	r *http.Request,
) error {
	const contentType = "application/vnd.contentful.management.v1+json"
	e := new(jx.Encoder)
	{
		req.Encode(e)
	}
	encoded := e.Bytes()
	ht.SetBody(r, bytes.NewReader(encoded), contentType)
	return nil
}

func encodePutAppInstallationRequest(
	req *PutAppInstallationReq,
	r *http.Request,
) error {
	const contentType = "application/vnd.contentful.management.v1+json"
	e := new(jx.Encoder)
	{
		req.Encode(e)
	}
	encoded := e.Bytes()
	ht.SetBody(r, bytes.NewReader(encoded), contentType)
	return nil
}

func encodePutContentTypeRequest(
	req *PutContentTypeReq,
	r *http.Request,
) error {
	const contentType = "application/vnd.contentful.management.v1+json"
	e := new(jx.Encoder)
	{
		req.Encode(e)
	}
	encoded := e.Bytes()
	ht.SetBody(r, bytes.NewReader(encoded), contentType)
	return nil
}

func encodePutEditorInterfaceRequest(
	req *PutEditorInterfaceReq,
	r *http.Request,
) error {
	const contentType = "application/vnd.contentful.management.v1+json"
	e := new(jx.Encoder)
	{
		req.Encode(e)
	}
	encoded := e.Bytes()
	ht.SetBody(r, bytes.NewReader(encoded), contentType)
	return nil
}

func encodeUpdateDeliveryApiKeyRequest(
	req *UpdateDeliveryApiKeyReq,
	r *http.Request,
) error {
	const contentType = "application/vnd.contentful.management.v1+json"
	e := new(jx.Encoder)
	{
		req.Encode(e)
	}
	encoded := e.Bytes()
	ht.SetBody(r, bytes.NewReader(encoded), contentType)
	return nil
}
