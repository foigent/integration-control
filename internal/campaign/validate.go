package campaign

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	digestPattern      = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	publicKeyIDPattern = regexp.MustCompile(`^ed25519:sha256:[0-9a-f]{64}$`)
	gitObjectPattern   = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
	portableIDPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]*$`)
	releaseVersion     = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9][A-Za-z0-9.-]*)?$`)
)

func ValidateManifest(value Manifest) error {
	if value.SchemaVersion != ManifestSchemaVersion ||
		value.Canonicalization != CanonicalizationVersion ||
		value.Profile != ManifestProfile {
		return errors.New("campaign manifest contract identity is invalid")
	}
	if !validPortableText(value.DecisionID, MaximumDecisionIdentifierRunes) ||
		!digestPattern.MatchString(value.DecisionSHA256) ||
		!validPortableText(value.CampaignID, MaximumCampaignIdentifierRunes) {
		return errors.New("campaign manifest decision or campaign identity is invalid")
	}
	frozenAt, err := parseCanonicalTime("manifest frozen_at", value.FrozenAt)
	if err != nil {
		return err
	}
	window, err := validateWindow(value.RequestedWindow)
	if err != nil {
		return fmt.Errorf("manifest requested window: %w", err)
	}
	if !frozenAt.Before(window.notAfter) {
		return errors.New("campaign manifest was frozen at or after its requested window end")
	}
	if !slices.Equal(value.RequestedActions, RequiredRequestedActions) {
		return errors.New("campaign manifest requested actions do not match the POC profile")
	}
	if !slices.Equal(value.ExplicitExclusions, RequiredExplicitExclusions) {
		return errors.New("campaign manifest explicit exclusions do not match the POC profile")
	}
	if err := validateArtifactReference(value.SourceBinding); err != nil {
		return fmt.Errorf("manifest Source binding: %w", err)
	}
	if value.SourceBinding.SchemaVersion != SourceBindingSchemaVersion {
		return errors.New("campaign manifest Source binding contract identity is invalid")
	}
	if value.RequestSet.MaxRequests != 20 || value.RequestSet.MaxFacts != 200 ||
		value.RequestSet.SchemaVersion != RequestSetSchemaVersion ||
		!digestPattern.MatchString(value.RequestSet.SHA256) {
		return errors.New("campaign manifest request-set binding is invalid")
	}
	if len(value.Components) != 2 {
		return errors.New("campaign manifest must bind exactly two product components")
	}
	expectedProducts := []string{"unirouter", "xinsight"}
	expectedRoles := []string{"source_evidence_producer", "source_evidence_consumer"}
	expectedContractIDs := []string{UniRouterComponentContractID, XInsightComponentContractID}
	for index, component := range value.Components {
		if component.Product != expectedProducts[index] || component.Role != expectedRoles[index] {
			return errors.New("campaign manifest components are not in the exact product-role order")
		}
		if component.Contract.ID != expectedContractIDs[index] {
			return fmt.Errorf("campaign manifest component %s contract identity is invalid", component.Product)
		}
		if err := validateContractReference(component.Contract); err != nil {
			return fmt.Errorf("manifest component %s contract: %w", component.Product, err)
		}
		if err := validateArtifactReference(component.Release); err != nil {
			return fmt.Errorf("manifest component %s release: %w", component.Product, err)
		}
		if err := validateArtifactReference(component.ScopeReceipt); err != nil {
			return fmt.Errorf("manifest component %s scope receipt: %w", component.Product, err)
		}
	}
	if err := validateSourceEvidenceSigning(value.SourceEvidenceSigning); err != nil {
		return err
	}
	if value.AuthorizationSigning.Purpose != AuthorizationPurpose ||
		value.AuthorizationSigning.Domain != AuthorizationDomain ||
		!digestPattern.MatchString(value.AuthorizationSigning.TrustSetSHA256) {
		return errors.New("campaign manifest authorization-signing authority is invalid")
	}
	if value.SourceEvidenceSigning.TrustSetSHA256 == value.AuthorizationSigning.TrustSetSHA256 {
		return errors.New("Source-evidence and authorization trust-set digests must be distinct")
	}
	return nil
}

func validateSourceEvidenceSigning(value SourceEvidenceSigning) error {
	if value.Purpose != SourceEvidencePurpose || value.Domain != SourceEvidenceDomain ||
		!validNamespacedReference(value.SigningKeyRef, "source-evidence-") ||
		!publicKeyIDPattern.MatchString(value.SignerPublicKeyID) ||
		!digestPattern.MatchString(value.TrustSetSHA256) ||
		value.ClaimScope != PackageClaimScope ||
		value.PackageScopeState != PackageScopeState ||
		value.NestedExecutionEvidenceV2 != NestedExecutionEvidenceV2State ||
		value.RuntimeCloseMarker {
		return errors.New("campaign manifest Source-evidence signing authority or completeness is invalid")
	}
	return nil
}

func ValidateAcknowledgement(value ComponentAcknowledgement) error {
	if value.SchemaVersion != AcknowledgementSchema ||
		(value.Product != "unirouter" && value.Product != "xinsight") ||
		!digestPattern.MatchString(value.CampaignManifestSHA256) ||
		value.Status != ComponentAcknowledgementAccepted {
		return errors.New("component acknowledgement identity or status is invalid")
	}
	if err := validateContractReference(value.ComponentContract); err != nil {
		return err
	}
	if err := validateArtifactReference(value.ComponentScopeReceipt); err != nil {
		return err
	}
	if err := validateArtifactReference(value.ProductVerificationReceipt); err != nil {
		return err
	}
	if err := validateContractReference(value.VerifierRelease); err != nil {
		return err
	}
	_, err := parseCanonicalTime("component acknowledgement verified_at", value.VerifiedAt)
	return err
}

func ValidateAcknowledgementAgainstManifest(value ComponentAcknowledgement, manifest Manifest) error {
	if err := ValidateAcknowledgement(value); err != nil {
		return err
	}
	manifestSHA, err := canonicalManifestSHA(manifest)
	if err != nil {
		return err
	}
	if value.CampaignManifestSHA256 != manifestSHA {
		return errors.New("component acknowledgement is bound to a different campaign manifest")
	}
	for _, component := range manifest.Components {
		if component.Product != value.Product {
			continue
		}
		if component.Contract != value.ComponentContract || component.ScopeReceipt != value.ComponentScopeReceipt {
			return errors.New("component acknowledgement contradicts its product manifest section")
		}
		return nil
	}
	return errors.New("component acknowledgement product is absent from the campaign manifest")
}

func ValidateAuthorization(value Authorization) error {
	if value.SchemaVersion != AuthorizationSchemaVersion ||
		value.Canonicalization != CanonicalizationVersion ||
		!validPortableText(value.AuthorizationID, MaximumAuthorizationIdentifierRune) ||
		!validPortableText(value.DecisionID, MaximumDecisionIdentifierRunes) ||
		!validPortableText(value.CampaignID, MaximumCampaignIdentifierRunes) ||
		!digestPattern.MatchString(value.CampaignManifestSHA256) ||
		!digestPattern.MatchString(value.ConfirmationSHA256) {
		return errors.New("campaign authorization identity is invalid")
	}
	if len(value.ComponentAcknowledgements) != 2 {
		return errors.New("campaign authorization must bind exactly two acknowledgements")
	}
	for index, expectedProduct := range []string{"unirouter", "xinsight"} {
		ack := value.ComponentAcknowledgements[index]
		if ack.Product != expectedProduct || ack.AcknowledgementSchemaVersion != AcknowledgementSchema ||
			!digestPattern.MatchString(ack.AcknowledgementSHA256) ||
			ack.CampaignManifestSHA256 != value.CampaignManifestSHA256 {
			return errors.New("campaign authorization acknowledgement references are invalid")
		}
	}
	if !slices.Equal(value.AuthorizedActions, RequiredRequestedActions) {
		return errors.New("campaign authorization actions do not match the POC profile")
	}
	window, err := validateWindow(value.AuthorizationWindow)
	if err != nil {
		return fmt.Errorf("authorization window: %w", err)
	}
	confirmedAt, err := parseCanonicalTime("authorization confirmed_at", value.ConfirmedAt)
	if err != nil {
		return err
	}
	if confirmedAt.Before(window.notBefore) || !confirmedAt.Before(window.notAfter) {
		return errors.New("authorization confirmation falls outside its window")
	}
	return nil
}

func ValidateAuthorizationAgainstManifest(value Authorization, manifest Manifest) error {
	if err := ValidateAuthorization(value); err != nil {
		return err
	}
	manifestSHA, err := canonicalManifestSHA(manifest)
	if err != nil {
		return err
	}
	if value.DecisionID != manifest.DecisionID || value.CampaignID != manifest.CampaignID ||
		value.CampaignManifestSHA256 != manifestSHA ||
		!slices.Equal(value.AuthorizedActions, manifest.RequestedActions) {
		return errors.New("campaign authorization does not match the preauthorization manifest")
	}
	authorized, _ := validateWindow(value.AuthorizationWindow)
	requested, _ := validateWindow(manifest.RequestedWindow)
	if authorized.notBefore.Before(requested.notBefore) || authorized.notAfter.After(requested.notAfter) {
		return errors.New("campaign authorization window exceeds the requested manifest window")
	}
	return nil
}

func canonicalManifestSHA(value Manifest) (string, error) {
	body, err := CanonicalManifest(value)
	if err != nil {
		return "", err
	}
	return SHA256(body), nil
}

func ValidateSignature(value AuthorizationSignature) error {
	if value.SchemaVersion != SignatureSchemaVersion || value.Purpose != AuthorizationPurpose ||
		value.Domain != AuthorizationDomain ||
		!validNamespacedReference(value.SigningKeyRef, "integration-authorization-") ||
		!publicKeyIDPattern.MatchString(value.SignerPublicKeyID) ||
		!digestPattern.MatchString(value.TrustSetSHA256) ||
		!digestPattern.MatchString(value.AuthorizationSHA256) ||
		value.Algorithm != SignatureAlgorithm {
		return errors.New("authorization signature envelope identity is invalid")
	}
	signature, err := base64.StdEncoding.DecodeString(value.Value)
	if err != nil || len(signature) != ed25519.SignatureSize || base64.StdEncoding.EncodeToString(signature) != value.Value {
		return errors.New("authorization signature value is not canonical Ed25519 base64")
	}
	return nil
}

func ValidateTrustSet(value AuthorizationTrustSet) error {
	if value.SchemaVersion != TrustSetSchemaVersion || value.Purpose != AuthorizationPurpose ||
		value.Domain != AuthorizationDomain ||
		!validNamespacedReference(value.TrustSetRef, "integration-authorization-trust-") ||
		len(value.Keys) == 0 || len(value.Keys) > MaximumAuthorizationTrustSetKeys {
		return errors.New("authorization trust-set identity is invalid")
	}
	seenRefs := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	seenKeys := map[string]struct{}{}
	previousID := ""
	for index, key := range value.Keys {
		if !validNamespacedReference(key.SigningKeyRef, "integration-authorization-") ||
			!publicKeyIDPattern.MatchString(key.SignerPublicKeyID) || key.Algorithm != SignatureAlgorithm ||
			(key.Status != AuthorizationTrustKeyActive && key.Status != AuthorizationTrustKeyDisabled) {
			return fmt.Errorf("authorization trust-set key[%d] identity is invalid", index)
		}
		publicKey, err := base64.StdEncoding.DecodeString(key.PublicKey)
		if err != nil || len(publicKey) != ed25519.PublicKeySize || base64.StdEncoding.EncodeToString(publicKey) != key.PublicKey {
			return fmt.Errorf("authorization trust-set key[%d] public key is invalid", index)
		}
		if PublicKeyID(publicKey) != key.SignerPublicKeyID {
			return fmt.Errorf("authorization trust-set key[%d] public-key identity mismatch", index)
		}
		window, err := validateWindow(AuthorizationWindow{NotBefore: key.ValidFrom, NotAfter: key.ValidUntil})
		if err != nil || !window.notBefore.Before(window.notAfter) {
			return fmt.Errorf("authorization trust-set key[%d] validity is invalid", index)
		}
		if previousID != "" && previousID >= key.SignerPublicKeyID {
			return errors.New("authorization trust-set keys are not in strict public-key identity order")
		}
		previousID = key.SignerPublicKeyID
		encodedKey := base64.StdEncoding.EncodeToString(publicKey)
		if _, duplicate := seenRefs[key.SigningKeyRef]; duplicate {
			return errors.New("authorization trust-set contains duplicate operational key references")
		}
		if _, duplicate := seenIDs[key.SignerPublicKeyID]; duplicate {
			return errors.New("authorization trust-set contains duplicate public-key identities")
		}
		if _, duplicate := seenKeys[encodedKey]; duplicate {
			return errors.New("authorization trust-set contains duplicate public-key bytes")
		}
		seenRefs[key.SigningKeyRef] = struct{}{}
		seenIDs[key.SignerPublicKeyID] = struct{}{}
		seenKeys[encodedKey] = struct{}{}
	}
	return nil
}

func PublicKeyID(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "ed25519:sha256:" + hex.EncodeToString(digest[:])
}

func ValidateRelease(value ReleaseManifest) error {
	if value.SchemaVersion != ReleaseSchemaVersion || value.Canonicalization != CanonicalizationVersion ||
		!releaseVersion.MatchString(value.Version) || !gitObjectPattern.MatchString(value.SourceCommit) ||
		!gitObjectPattern.MatchString(value.SourceTree) || value.GoVersion != "go1.25.0" ||
		value.Toolchain != "go1.25.12" || !digestPattern.MatchString(value.ContractsSHA256) ||
		!digestPattern.MatchString(value.ValidFixturesSHA256) || !digestPattern.MatchString(value.InvalidFixturesSHA256) ||
		!value.DoubleBuildEqual || value.CapabilityBoundary != CapabilityBoundaryOfflineOnly {
		return errors.New("integration-control release identity is invalid")
	}
	if len(value.Artifacts) == 0 || len(value.Gates) == 0 {
		return errors.New("integration-control release evidence is incomplete")
	}
	lastArtifact := ""
	for _, artifact := range value.Artifacts {
		if !validPathText(artifact.Path) || artifact.Path <= lastArtifact || artifact.SizeBytes <= 0 ||
			!digestPattern.MatchString(artifact.SHA256) || !validPortableText(artifact.OS, 40) ||
			!validPortableText(artifact.Architecture, 40) {
			return errors.New("integration-control release artifact identity is invalid or unsorted")
		}
		lastArtifact = artifact.Path
	}
	lastGate := ""
	for _, gate := range value.Gates {
		if !validPortableText(gate.Name, 120) || gate.Name <= lastGate || gate.Status != ReleaseGatePass {
			return errors.New("integration-control release gates are invalid or unsorted")
		}
		lastGate = gate.Name
	}
	if value.Effects != (ReleaseEffects{}) {
		return errors.New("integration-control release declares external effects")
	}
	return nil
}

func validateArtifactReference(value ArtifactReference) error {
	if !validSchemaVersion(value.SchemaVersion) || !digestPattern.MatchString(value.SHA256) {
		return errors.New("artifact reference is invalid")
	}
	return nil
}

func validateContractReference(value ContractReference) error {
	if !validPortableText(value.ID, MaximumContractIdentifierRunes) || !digestPattern.MatchString(value.SHA256) {
		return errors.New("contract reference is invalid")
	}
	return nil
}

type validatedWindow struct {
	notBefore time.Time
	notAfter  time.Time
}

func validateWindow(value AuthorizationWindow) (validatedWindow, error) {
	notBefore, err := parseCanonicalTime("window not_before", value.NotBefore)
	if err != nil {
		return validatedWindow{}, err
	}
	notAfter, err := parseCanonicalTime("window not_after", value.NotAfter)
	if err != nil {
		return validatedWindow{}, err
	}
	if !notBefore.Before(notAfter) {
		return validatedWindow{}, errors.New("window not_before must precede not_after")
	}
	return validatedWindow{notBefore: notBefore, notAfter: notAfter}, nil
}

func parseCanonicalTime(label, value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC || parsed.Format(time.RFC3339Nano) != value || !strings.HasSuffix(value, "Z") {
		return time.Time{}, fmt.Errorf("%s is not canonical UTC RFC3339", label)
	}
	return parsed, nil
}

func validSchemaVersion(value string) bool {
	return validPortableText(value, MaximumArtifactSchemaVersionRunes)
}

func validNamespacedReference(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && validPortableText(value, MaximumOperationalReferenceRunes)
}

func validPortableText(value string, maximum int) bool {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > maximum ||
		value != strings.TrimSpace(value) || !portableIDPattern.MatchString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validPathText(value string) bool {
	return validPortableText(value, MaximumReleaseArtifactPathRunes) && !strings.Contains(value, "..") && !strings.HasPrefix(value, "/")
}
