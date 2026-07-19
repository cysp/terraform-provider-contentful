package contentfulmanagement

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/go-faster/jx"
)

var errNonCanonicalAppKeyJWKMaterial = errors.New("x5c is not canonical standard base64")

type AppKeyJWKMaterial struct {
	DER         []byte
	Fingerprint string
}

func DecodeAppKeyJWKMaterial(x5c string) (AppKeyJWKMaterial, error) {
	publicKeyDER, err := base64.StdEncoding.DecodeString(x5c)
	if err != nil {
		return AppKeyJWKMaterial{}, fmt.Errorf("decode standard base64: %w", err)
	}

	if base64.StdEncoding.EncodeToString(publicKeyDER) != x5c {
		return AppKeyJWKMaterial{}, errNonCanonicalAppKeyJWKMaterial
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
