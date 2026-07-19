package contentfulmanagement

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

var errAppKeyJWKBase64LineBreak = errors.New("CR and LF are not permitted")

func AppKeyJWKFingerprintFromX5C(x5c string) (string, error) {
	if strings.ContainsAny(x5c, "\r\n") {
		return "", fmt.Errorf("decode standard base64: %w", errAppKeyJWKBase64LineBreak)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(x5c)
	if err != nil {
		return "", fmt.Errorf("decode standard base64: %w", err)
	}

	return AppKeyJWKFingerprint(publicKeyBytes), nil
}

func AppKeyJWKFingerprint(publicKeyBytes []byte) string {
	fingerprint := sha256.Sum256(publicKeyBytes)

	return base64.RawURLEncoding.EncodeToString(fingerprint[:])
}

func NewAppKeyRequestData(jwk AppKeyJWK) AppKeyRequestData {
	return AppKeyRequestData{
		Jwk: jwk,
	}
}
