package cmtesting

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
)

const (
	appKeyDSAParameterBits = 2048
	appKeyDSASubprimeBits  = 256
	bitsPerByte            = 8
)

// NewDSAAppKeyPublicKeyDER returns deterministic, structurally valid DSA
// SubjectPublicKeyInfo for testing Contentful's opaque x5c handling.
func NewDSAAppKeyPublicKeyDER() ([]byte, error) {
	prime := appKeyLargeInteger(appKeyDSAParameterBits)
	subprime := appKeyLargeInteger(appKeyDSASubprimeBits)
	generator := appKeyLargeInteger(appKeyDSAParameterBits)
	publicValue := appKeyLargeInteger(appKeyDSAParameterBits)

	parameters, err := asn1.Marshal(struct {
		P *big.Int
		Q *big.Int
		G *big.Int
	}{prime, subprime, generator})
	if err != nil {
		return nil, fmt.Errorf("marshal DSA parameters: %w", err)
	}

	encodedPublicValue, err := asn1.Marshal(publicValue)
	if err != nil {
		return nil, fmt.Errorf("marshal DSA public value: %w", err)
	}

	publicKeyDER, err := asn1.Marshal(struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm:  asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 1},
			Parameters: asn1.RawValue{FullBytes: parameters},
		},
		PublicKey: asn1.BitString{Bytes: encodedPublicValue, BitLength: len(encodedPublicValue) * bitsPerByte},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal DSA SubjectPublicKeyInfo: %w", err)
	}

	return publicKeyDER, nil
}

func appKeyLargeInteger(bits uint) *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(1))
}
