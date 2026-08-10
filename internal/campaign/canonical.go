package campaign

import (
	"bytes"
	"encoding/json"
	"errors"
)

func canonicalBytes(value any) ([]byte, error) {
	return json.Marshal(value)
}

func DecodeManifest(body []byte) (Manifest, error) {
	value, err := decodeStrict[Manifest](body)
	if err != nil {
		return Manifest{}, err
	}
	if err := ValidateManifest(value); err != nil {
		return Manifest{}, err
	}
	return value, nil
}

func CanonicalManifest(value Manifest) ([]byte, error) {
	if err := ValidateManifest(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalManifest(body []byte) (Manifest, error) {
	value, err := DecodeManifest(body)
	if err != nil {
		return Manifest{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return Manifest{}, errors.New("campaign manifest is not canonical JSON")
	}
	return value, nil
}

func DecodeAcknowledgement(body []byte) (ComponentAcknowledgement, error) {
	value, err := decodeStrict[ComponentAcknowledgement](body)
	if err != nil {
		return ComponentAcknowledgement{}, err
	}
	if err := ValidateAcknowledgement(value); err != nil {
		return ComponentAcknowledgement{}, err
	}
	return value, nil
}

func CanonicalAcknowledgement(value ComponentAcknowledgement) ([]byte, error) {
	if err := ValidateAcknowledgement(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalAcknowledgement(body []byte) (ComponentAcknowledgement, error) {
	value, err := DecodeAcknowledgement(body)
	if err != nil {
		return ComponentAcknowledgement{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return ComponentAcknowledgement{}, errors.New("component acknowledgement is not canonical JSON")
	}
	return value, nil
}

func DecodeAuthorization(body []byte) (Authorization, error) {
	value, err := decodeStrict[Authorization](body)
	if err != nil {
		return Authorization{}, err
	}
	if err := ValidateAuthorization(value); err != nil {
		return Authorization{}, err
	}
	return value, nil
}

func CanonicalAuthorization(value Authorization) ([]byte, error) {
	if err := ValidateAuthorization(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalAuthorization(body []byte) (Authorization, error) {
	value, err := DecodeAuthorization(body)
	if err != nil {
		return Authorization{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return Authorization{}, errors.New("campaign authorization is not canonical JSON")
	}
	return value, nil
}

func DecodeSignature(body []byte) (AuthorizationSignature, error) {
	value, err := decodeStrict[AuthorizationSignature](body)
	if err != nil {
		return AuthorizationSignature{}, err
	}
	if err := ValidateSignature(value); err != nil {
		return AuthorizationSignature{}, err
	}
	return value, nil
}

func CanonicalSignature(value AuthorizationSignature) ([]byte, error) {
	if err := ValidateSignature(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalSignature(body []byte) (AuthorizationSignature, error) {
	value, err := DecodeSignature(body)
	if err != nil {
		return AuthorizationSignature{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return AuthorizationSignature{}, errors.New("authorization signature envelope is not canonical JSON")
	}
	return value, nil
}

func DecodeTrustSet(body []byte) (AuthorizationTrustSet, error) {
	value, err := decodeStrict[AuthorizationTrustSet](body)
	if err != nil {
		return AuthorizationTrustSet{}, err
	}
	if err := ValidateTrustSet(value); err != nil {
		return AuthorizationTrustSet{}, err
	}
	return value, nil
}

func CanonicalTrustSet(value AuthorizationTrustSet) ([]byte, error) {
	if err := ValidateTrustSet(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalTrustSet(body []byte) (AuthorizationTrustSet, error) {
	value, err := DecodeTrustSet(body)
	if err != nil {
		return AuthorizationTrustSet{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return AuthorizationTrustSet{}, errors.New("authorization trust set is not canonical JSON")
	}
	return value, nil
}

func DecodeRelease(body []byte) (ReleaseManifest, error) {
	value, err := decodeStrict[ReleaseManifest](body)
	if err != nil {
		return ReleaseManifest{}, err
	}
	if err := ValidateRelease(value); err != nil {
		return ReleaseManifest{}, err
	}
	return value, nil
}

func CanonicalRelease(value ReleaseManifest) ([]byte, error) {
	if err := ValidateRelease(value); err != nil {
		return nil, err
	}
	return canonicalBytes(value)
}

func DecodeCanonicalRelease(body []byte) (ReleaseManifest, error) {
	value, err := DecodeRelease(body)
	if err != nil {
		return ReleaseManifest{}, err
	}
	canonical, _ := canonicalBytes(value)
	if !bytes.Equal(body, canonical) {
		return ReleaseManifest{}, errors.New("integration-control release manifest is not canonical JSON")
	}
	return value, nil
}
