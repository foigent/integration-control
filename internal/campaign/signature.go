package campaign

import (
	"bytes"
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

const AuthorizationPreimagePrefix = AuthorizationDomain + "\x00" + AuthorizationPurpose + "\x00"

func AuthorizationPreimage(canonicalAuthorization []byte) []byte {
	preimage := make([]byte, 0, len(AuthorizationPreimagePrefix)+len(canonicalAuthorization))
	preimage = append(preimage, AuthorizationPreimagePrefix...)
	preimage = append(preimage, canonicalAuthorization...)
	return preimage
}

// VerifyAuthorizationSignature selects exactly one purpose-specific public key
// from the immutable authorization trust set. It never consults system keyrings,
// product trust stores, or network services.
func VerifyAuthorizationSignature(manifest Manifest, authorization Authorization, canonicalAuthorization []byte, signature AuthorizationSignature, trustSet AuthorizationTrustSet, canonicalTrustSet []byte) error {
	if err := ValidateManifest(manifest); err != nil {
		return err
	}
	if err := ValidateAuthorizationAgainstManifest(authorization, manifest); err != nil {
		return err
	}
	if err := ValidateSignature(signature); err != nil {
		return err
	}
	if err := ValidateTrustSet(trustSet); err != nil {
		return err
	}
	expectedAuthorization, _ := CanonicalAuthorization(authorization)
	if !bytes.Equal(expectedAuthorization, canonicalAuthorization) {
		return errors.New("authorization value and canonical bytes disagree")
	}
	expectedTrustSet, _ := CanonicalTrustSet(trustSet)
	if !bytes.Equal(expectedTrustSet, canonicalTrustSet) {
		return errors.New("authorization trust-set value and canonical bytes disagree")
	}
	if SHA256(canonicalAuthorization) != signature.AuthorizationSHA256 {
		return errors.New("authorization signature does not bind the exact canonical authorization bytes")
	}
	authorizationTrustSHA := SHA256(canonicalTrustSet)
	if manifest.AuthorizationSigning.TrustSetSHA256 != authorizationTrustSHA || signature.TrustSetSHA256 != authorizationTrustSHA {
		return errors.New("authorization signature does not bind the manifest authorization trust set")
	}
	if subtle.ConstantTimeCompare([]byte(authorizationTrustSHA), []byte(manifest.SourceEvidenceSigning.TrustSetSHA256)) == 1 {
		return errors.New("authorization and Source-evidence trust authorities are conflated")
	}

	var selected *AuthorizationTrustKey
	for index := range trustSet.Keys {
		key := &trustSet.Keys[index]
		if key.SigningKeyRef == signature.SigningKeyRef && key.SignerPublicKeyID == signature.SignerPublicKeyID {
			if selected != nil {
				return errors.New("authorization trust selection is ambiguous")
			}
			selected = key
		}
	}
	if selected == nil {
		return errors.New("authorization signing key is not present in the selected trust set")
	}
	if selected.Status != AuthorizationTrustKeyActive || selected.Algorithm != signature.Algorithm {
		return errors.New("authorization signing key is not active for the required algorithm")
	}
	confirmedAt, _ := parseCanonicalTime("authorization confirmed_at", authorization.ConfirmedAt)
	validFrom, _ := time.Parse(time.RFC3339Nano, selected.ValidFrom)
	validUntil, _ := time.Parse(time.RFC3339Nano, selected.ValidUntil)
	if confirmedAt.Before(validFrom) || !confirmedAt.Before(validUntil) {
		return errors.New("authorization was confirmed outside the selected key validity window")
	}
	publicKey, _ := base64.StdEncoding.DecodeString(selected.PublicKey)
	signatureBytes, _ := base64.StdEncoding.DecodeString(signature.Value)
	if !ed25519.Verify(ed25519.PublicKey(publicKey), AuthorizationPreimage(canonicalAuthorization), signatureBytes) {
		return errors.New("authorization Ed25519 signature verification failed")
	}
	return nil
}

type WholeVerification struct {
	Valid                        bool              `json:"valid"`
	CampaignManifestSHA256       string            `json:"campaign_manifest_sha256"`
	AcknowledgementSHA256        map[string]string `json:"acknowledgement_sha256"`
	AuthorizationSHA256          string            `json:"authorization_sha256"`
	AuthorizationSignatureSHA256 string            `json:"authorization_signature_sha256"`
	AuthorizationTrustSetSHA256  string            `json:"authorization_trust_set_sha256"`
	SourceEvidenceTrustSetSHA256 string            `json:"source_evidence_trust_set_sha256"`
	CapabilityBoundary           string            `json:"capability_boundary"`
}

// VerifyWhole verifies only neutral artifact bytes and references. Product
// verification receipts remain opaque product-owned facts.
func VerifyWhole(manifestBody, uniRouterAckBody, xInsightAckBody, authorizationBody, signatureBody, trustSetBody []byte) (WholeVerification, error) {
	manifest, err := DecodeCanonicalManifest(manifestBody)
	if err != nil {
		return WholeVerification{}, fmt.Errorf("manifest: %w", err)
	}
	manifestSHA := SHA256(manifestBody)
	ackBodies := [][]byte{uniRouterAckBody, xInsightAckBody}
	expectedProducts := []string{"unirouter", "xinsight"}
	ackDigests := make(map[string]string, 2)
	for index, body := range ackBodies {
		acknowledgement, ackErr := DecodeCanonicalAcknowledgement(body)
		if ackErr != nil {
			return WholeVerification{}, fmt.Errorf("acknowledgement[%d]: %w", index, ackErr)
		}
		if ackErr = ValidateAcknowledgementAgainstManifest(acknowledgement, manifest); ackErr != nil {
			return WholeVerification{}, fmt.Errorf("acknowledgement[%d]: %w", index, ackErr)
		}
		if acknowledgement.Product != expectedProducts[index] {
			return WholeVerification{}, fmt.Errorf("acknowledgement[%d] is not the expected %s component", index, expectedProducts[index])
		}
		if _, duplicate := ackDigests[acknowledgement.Product]; duplicate {
			return WholeVerification{}, errors.New("duplicate product acknowledgement")
		}
		ackDigests[acknowledgement.Product] = SHA256(body)
	}
	if len(ackDigests) != 2 || ackDigests["unirouter"] == "" || ackDigests["xinsight"] == "" {
		return WholeVerification{}, errors.New("both product acknowledgements are required")
	}

	authorization, err := DecodeCanonicalAuthorization(authorizationBody)
	if err != nil {
		return WholeVerification{}, fmt.Errorf("authorization: %w", err)
	}
	if err = ValidateAuthorizationAgainstManifest(authorization, manifest); err != nil {
		return WholeVerification{}, err
	}
	for _, reference := range authorization.ComponentAcknowledgements {
		if ackDigests[reference.Product] != reference.AcknowledgementSHA256 {
			return WholeVerification{}, fmt.Errorf("authorization acknowledgement digest mismatch for %s", reference.Product)
		}
	}
	if err = validateAcknowledgementChronology(manifest, authorization, ackBodies); err != nil {
		return WholeVerification{}, err
	}
	signature, err := DecodeCanonicalSignature(signatureBody)
	if err != nil {
		return WholeVerification{}, fmt.Errorf("signature: %w", err)
	}
	trustSet, err := DecodeCanonicalTrustSet(trustSetBody)
	if err != nil {
		return WholeVerification{}, fmt.Errorf("trust set: %w", err)
	}
	if err = VerifyAuthorizationSignature(manifest, authorization, authorizationBody, signature, trustSet, trustSetBody); err != nil {
		return WholeVerification{}, err
	}
	return WholeVerification{
		Valid:                        true,
		CampaignManifestSHA256:       manifestSHA,
		AcknowledgementSHA256:        ackDigests,
		AuthorizationSHA256:          SHA256(authorizationBody),
		AuthorizationSignatureSHA256: SHA256(signatureBody),
		AuthorizationTrustSetSHA256:  SHA256(trustSetBody),
		SourceEvidenceTrustSetSHA256: manifest.SourceEvidenceSigning.TrustSetSHA256,
		CapabilityBoundary:           CapabilityBoundaryOfflineOnly,
	}, nil
}
