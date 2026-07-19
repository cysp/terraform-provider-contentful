package contentfulmanagement

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-faster/jx"
)

var errAppKeyJWKBase64LineBreak = errors.New("CR and LF are not permitted")

type AppKeyJWKMaterial struct {
	DER         []byte
	Fingerprint string
}

func DecodeAppKeyJWKMaterial(x5c string) (AppKeyJWKMaterial, error) {
	if strings.ContainsAny(x5c, "\r\n") {
		return AppKeyJWKMaterial{}, fmt.Errorf("decode standard base64: %w", errAppKeyJWKBase64LineBreak)
	}

	publicKeyDER, err := base64.StdEncoding.DecodeString(x5c)
	if err != nil {
		return AppKeyJWKMaterial{}, fmt.Errorf("decode standard base64: %w", err)
	}

	return AppKeyJWKMaterial{
		DER:         publicKeyDER,
		Fingerprint: AppKeyJWKFingerprint(publicKeyDER),
	}, nil
}

func AppKeyJWKFingerprint(publicKeyDER []byte) string {
	fingerprint := sha256.Sum256(publicKeyDER)

	return base64.RawURLEncoding.EncodeToString(fingerprint[:])
}

func NewAppKeyRequestData(jwk AppKeyJWK) AppKeyRequestData {
	encoder := jx.Encoder{}
	jwk.Encode(&encoder)

	return AppKeyRequestData{
		Jwk: encoder.Bytes(),
	}
}
