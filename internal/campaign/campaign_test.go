package campaign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "v1", "valid", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func invalidFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "v1", "invalid", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestValidFixtureWholeVerification(t *testing.T) {
	report, err := VerifyWhole(
		fixture(t, "manifest.json"),
		fixture(t, "unirouter-ack.json"),
		fixture(t, "xinsight-ack.json"),
		fixture(t, "authorization.json"),
		fixture(t, "authorization-signature.json"),
		fixture(t, "authorization-trust-set.json"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Valid || report.CapabilityBoundary != CapabilityBoundaryOfflineOnly {
		t.Fatalf("unexpected verification report: %#v", report)
	}
	manifest, err := DecodeCanonicalManifest(fixture(t, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	trust, err := DecodeCanonicalTrustSet(fixture(t, "authorization-trust-set.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.SourceEvidenceSigning.SignerPublicKeyID != trust.Keys[0].SignerPublicKeyID {
		t.Fatal("fixture must exercise same-key-bytes purpose separation")
	}
	var sourceTrust map[string]any
	if err := json.Unmarshal(fixture(t, "source-evidence-trust-set.json"), &sourceTrust); err != nil {
		t.Fatal(err)
	}
	if sourceTrust["public_key"] != trust.Keys[0].PublicKey {
		t.Fatal("fixture must exercise identical public-key bytes under separate purposes")
	}
	if report.AuthorizationTrustSetSHA256 == report.SourceEvidenceTrustSetSHA256 {
		t.Fatal("purpose-specific trust-set digests were conflated")
	}
}

func TestUniRouterFixtureUsesAcceptedSourceScopeAndReleaseAuthority(t *testing.T) {
	var source map[string]any
	if err := json.Unmarshal(fixture(t, "source-binding.json"), &source); err != nil {
		t.Fatal(err)
	}
	expectedSource := map[string]any{
		"schema_version":       "unirouter.source-evidence.source-scope.v1",
		"product":              "unirouter",
		"source_instance":      "fixture-source-20260811",
		"identity_authority":   "neutral_campaign_authorization",
		"runtime_observation":  "not_claimed",
		"feed_snapshot_schema": "unirouter.completed-execution.v1",
	}
	if !reflect.DeepEqual(source, expectedSource) {
		t.Fatalf("Source scope fixture claims unaccepted runtime facts: %#v", source)
	}
	var requestSet map[string]any
	if err := json.Unmarshal(fixture(t, "request-set.json"), &requestSet); err != nil {
		t.Fatal(err)
	}
	if requestSet["source_instance"] != source["source_instance"] {
		t.Fatal("Source scope source_instance does not come from the frozen request set")
	}
	if requestSet["schema_version"] != RequestSetSchemaVersion {
		t.Fatalf("request-set fixture uses non-normative identity %v", requestSet["schema_version"])
	}
	manifest, err := DecodeCanonicalManifest(fixture(t, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.SourceBinding.SchemaVersion != SourceBindingSchemaVersion ||
		manifest.RequestSet.SchemaVersion != RequestSetSchemaVersion ||
		manifest.Components[0].Contract.ID != UniRouterComponentContractID {
		t.Fatal("positive manifest does not use the fixed POC profile contract identities")
	}
	var release map[string]any
	if err := json.Unmarshal(fixture(t, "unirouter-release.json"), &release); err != nil {
		t.Fatal(err)
	}
	if _, claimsTree := release["tree"]; claimsTree {
		t.Fatal("UniRouter installed release fixture must not claim a runtime tree")
	}
	for _, required := range []string{"sequence", "version", "canonical_commit", "package_sha256", "signing_key_id", "database_schema", "observed_at"} {
		if release[required] == nil {
			t.Fatalf("UniRouter installed release fixture is missing %s", required)
		}
	}
}

func TestManifestFixedProfileRejectsWrongContractIdentities(t *testing.T) {
	original, err := DecodeCanonicalManifest(fixture(t, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"wrong Source binding identity", func(value *Manifest) {
			value.SourceBinding.SchemaVersion = "unirouter.source-evidence.source-scope.v2"
		}},
		{"wrong request-set identity", func(value *Manifest) {
			value.RequestSet.SchemaVersion = "foigent.poc-request-set.v1"
		}},
		{"wrong UniRouter component contract identity", func(value *Manifest) {
			value.Components[0].Contract.ID = "unirouter.source-evidence.component.v1"
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mutated := original
			mutated.Components = append([]ComponentInput(nil), original.Components...)
			testCase.mutate(&mutated)
			if err := ValidateManifest(mutated); err == nil {
				t.Fatal("fixed POC profile accepted a widened contract identity")
			}
		})
	}
}

func TestManifestBuilderHashesExactInputs(t *testing.T) {
	inputs := validManifestInputs(t)
	_, body, err := BuildManifestFromDraft(fixture(t, "manifest-draft.json"), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, fixture(t, "manifest.json")) {
		t.Fatal("manifest fixture is not reproducible from exact input bytes")
	}

	inputs.SourceBinding = append(append([]byte(nil), inputs.SourceBinding...), ' ')
	_, changed, err := BuildManifestFromDraft(fixture(t, "manifest-draft.json"), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(body, changed) {
		t.Fatal("exact-byte Source-binding change did not change the manifest")
	}
}

func validManifestInputs(t *testing.T) ManifestBuildInputs {
	t.Helper()
	return ManifestBuildInputs{
		Decision: fixture(t, "decision.txt"), SourceBinding: fixture(t, "source-binding.json"),
		RequestSet: fixture(t, "request-set.json"), SourceEvidenceTrustSet: fixture(t, "source-evidence-trust-set.json"),
		AuthorizationTrustSet: fixture(t, "authorization-trust-set.json"), UniRouterContract: fixture(t, "unirouter-contract.json"),
		UniRouterRelease: fixture(t, "unirouter-release.json"), UniRouterScope: fixture(t, "unirouter-scope.json"),
		XInsightContract: fixture(t, "xinsight-contract.json"), XInsightRelease: fixture(t, "xinsight-release.json"),
		XInsightScope: fixture(t, "xinsight-scope.json"),
	}
}

func TestManifestBuilderRejectsOperatorEnteredDigests(t *testing.T) {
	draft, err := decodeStrict[Manifest](fixture(t, "manifest-draft.json"))
	if err != nil {
		t.Fatal(err)
	}
	draft.DecisionSHA256 = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body, _ := json.Marshal(draft)
	_, _, err = BuildManifestFromDraft(body, ManifestBuildInputs{})
	if err == nil || !strings.Contains(err.Error(), "pre-filled") {
		t.Fatalf("expected pre-filled digest rejection, got %v", err)
	}
}

func TestAuthorizationBuilderDoesNotPersistConfirmationText(t *testing.T) {
	confirmation := []byte("synthetic-test-only-confirmation-material")
	_, body, err := BuildAuthorizationFromDraft(
		fixture(t, "authorization-draft.json"), fixture(t, "manifest.json"),
		fixture(t, "unirouter-ack.json"), fixture(t, "xinsight-ack.json"), confirmation,
	)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, confirmation) {
		t.Fatal("authorization persisted raw confirmation material")
	}
	authorization, err := DecodeCanonicalAuthorization(body)
	if err != nil {
		t.Fatal(err)
	}
	if authorization.ConfirmationSHA256 != SHA256(confirmation) {
		t.Fatal("authorization did not bind exact confirmation bytes")
	}
}

func TestAcknowledgementBoundaryDoesNotAssertProductReceiptContents(t *testing.T) {
	manifestBody := fixture(t, "manifest.json")
	manifest, _ := DecodeCanonicalManifest(manifestBody)
	ack, _ := DecodeCanonicalAcknowledgement(fixture(t, "unirouter-ack.json"))
	ack.ProductVerificationReceipt = ArtifactReference{
		SchemaVersion: "unirouter.product-owned-verification.v99",
		SHA256:        "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}
	if err := ValidateAcknowledgementAgainstManifest(ack, manifest); err != nil {
		t.Fatalf("neutral verifier interpreted product receipt contents: %v", err)
	}
	ack.Status = "rejected"
	if err := ValidateAcknowledgementAgainstManifest(ack, manifest); err == nil {
		t.Fatal("neutral verifier accepted a non-accepted product status")
	}
}

func TestWholeVerificationRejectsReversedAcknowledgements(t *testing.T) {
	_, err := VerifyWhole(
		fixture(t, "manifest.json"), fixture(t, "xinsight-ack.json"), fixture(t, "unirouter-ack.json"),
		fixture(t, "authorization.json"), fixture(t, "authorization-signature.json"), fixture(t, "authorization-trust-set.json"),
	)
	if err == nil || !strings.Contains(err.Error(), "expected unirouter") {
		t.Fatalf("expected ordered acknowledgement rejection, got %v", err)
	}
}

func TestAuthorizationBuilderRejectsPostConfirmationAcknowledgement(t *testing.T) {
	ack, err := DecodeCanonicalAcknowledgement(fixture(t, "unirouter-ack.json"))
	if err != nil {
		t.Fatal(err)
	}
	ack.VerifiedAt = "2030-01-01T00:00:01Z"
	ackBody, err := CanonicalAcknowledgement(ack)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = BuildAuthorizationFromDraft(
		fixture(t, "authorization-draft.json"), fixture(t, "manifest.json"),
		ackBody, fixture(t, "xinsight-ack.json"), []byte("test-only"),
	)
	if err == nil || !strings.Contains(err.Error(), "outside manifest-freeze to user-confirmation") {
		t.Fatalf("expected acknowledgement chronology rejection, got %v", err)
	}
}

func TestPurposeSeparatedTrustNegativeMatrix(t *testing.T) {
	manifestBody := fixture(t, "manifest.json")
	manifest, _ := DecodeCanonicalManifest(manifestBody)
	authorizationBody := fixture(t, "authorization.json")
	authorization, _ := DecodeCanonicalAuthorization(authorizationBody)
	signature, _ := DecodeCanonicalSignature(fixture(t, "authorization-signature.json"))
	trustBody := fixture(t, "authorization-trust-set.json")
	trust, _ := DecodeCanonicalTrustSet(trustBody)

	t.Run("cross-purpose trust digest cannot occupy both manifest fields", func(t *testing.T) {
		mutated := manifest
		mutated.SourceEvidenceSigning.TrustSetSHA256 = mutated.AuthorizationSigning.TrustSetSHA256
		if err := ValidateManifest(mutated); err == nil || !strings.Contains(err.Error(), "must be distinct") {
			t.Fatalf("expected trust digest conflation rejection, got %v", err)
		}
	})

	t.Run("source trust digest cannot verify authorization", func(t *testing.T) {
		mutated := signature
		mutated.TrustSetSHA256 = manifest.SourceEvidenceSigning.TrustSetSHA256
		if err := VerifyAuthorizationSignature(manifest, authorization, authorizationBody, mutated, trust, trustBody); err == nil {
			t.Fatal("Source-evidence trust digest verified an authorization")
		}
	})

	t.Run("source operational key reference cannot verify authorization", func(t *testing.T) {
		mutated := signature
		mutated.SigningKeyRef = manifest.SourceEvidenceSigning.SigningKeyRef
		if err := VerifyAuthorizationSignature(manifest, authorization, authorizationBody, mutated, trust, trustBody); err == nil {
			t.Fatal("Source-evidence operational key reference verified an authorization")
		}
	})

	for _, authority := range []struct{ name, purpose, domain string }{
		{name: "source evidence authority", purpose: SourceEvidencePurpose, domain: SourceEvidenceDomain},
		{name: "upgrade authority", purpose: "upgrade", domain: "unirouter.release-upgrade.v1"},
		{name: "release authority", purpose: "release", domain: "foigent.release.v1"},
	} {
		t.Run(authority.name+" cannot become authorization trust", func(t *testing.T) {
			mutated := trust
			mutated.Purpose = authority.purpose
			mutated.Domain = authority.domain
			if err := ValidateTrustSet(mutated); err == nil {
				t.Fatalf("%s was accepted as authorization trust", authority.name)
			}
		})
	}

	t.Run("disabled key cannot verify", func(t *testing.T) {
		mutated := trust
		mutated.Keys = append([]AuthorizationTrustKey(nil), trust.Keys...)
		mutated.Keys[0].Status = AuthorizationTrustKeyDisabled
		mutatedBody, _ := canonicalBytes(mutated)
		mutatedManifest := manifest
		mutatedManifest.AuthorizationSigning.TrustSetSHA256 = SHA256(mutatedBody)
		mutatedSignature := signature
		mutatedSignature.TrustSetSHA256 = SHA256(mutatedBody)
		if err := VerifyAuthorizationSignature(mutatedManifest, authorization, authorizationBody, mutatedSignature, mutated, mutatedBody); err == nil {
			t.Fatal("disabled authorization key was accepted")
		}
	})
}

func TestSignatureAndBindingTamperMatrix(t *testing.T) {
	manifestBody := fixture(t, "manifest.json")
	manifest, _ := DecodeCanonicalManifest(manifestBody)
	authorizationBody := fixture(t, "authorization.json")
	authorization, _ := DecodeCanonicalAuthorization(authorizationBody)
	signature, _ := DecodeCanonicalSignature(fixture(t, "authorization-signature.json"))
	trustBody := fixture(t, "authorization-trust-set.json")
	trust, _ := DecodeCanonicalTrustSet(trustBody)

	t.Run("authorization content tamper", func(t *testing.T) {
		mutated := authorization
		mutated.ConfirmationSHA256 = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		mutatedBody, _ := CanonicalAuthorization(mutated)
		if err := VerifyAuthorizationSignature(manifest, mutated, mutatedBody, signature, trust, trustBody); err == nil {
			t.Fatal("tampered authorization was accepted")
		}
	})

	t.Run("authorization value and byte disagreement", func(t *testing.T) {
		other := authorization
		other.ConfirmationSHA256 = "sha256:abababababababababababababababababababababababababababababababab"
		otherBody, _ := CanonicalAuthorization(other)
		err := VerifyAuthorizationSignature(manifest, authorization, otherBody, signature, trust, trustBody)
		if err == nil || !strings.Contains(err.Error(), "value and canonical bytes disagree") {
			t.Fatalf("expected authorization value/bytes rejection, got %v", err)
		}
	})

	t.Run("trust-set value and byte disagreement", func(t *testing.T) {
		other := trust
		other.Keys = append([]AuthorizationTrustKey(nil), trust.Keys...)
		other.Keys[0].Status = AuthorizationTrustKeyDisabled
		otherBody, _ := CanonicalTrustSet(other)
		err := VerifyAuthorizationSignature(manifest, authorization, authorizationBody, signature, trust, otherBody)
		if err == nil || !strings.Contains(err.Error(), "value and canonical bytes disagree") {
			t.Fatalf("expected trust-set value/bytes rejection, got %v", err)
		}
	})

	t.Run("signature value tamper", func(t *testing.T) {
		mutated := signature
		value := []byte(mutated.Value)
		if value[0] == 'A' {
			value[0] = 'B'
		} else {
			value[0] = 'A'
		}
		mutated.Value = string(value)
		if err := VerifyAuthorizationSignature(manifest, authorization, authorizationBody, mutated, trust, trustBody); err == nil {
			t.Fatal("tampered signature was accepted")
		}
	})

	t.Run("acknowledgement digest tamper", func(t *testing.T) {
		mutated := authorization
		mutated.ComponentAcknowledgements = append([]ComponentAcknowledgementReference(nil), authorization.ComponentAcknowledgements...)
		mutated.ComponentAcknowledgements[0].AcknowledgementSHA256 = "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
		mutatedBody, _ := CanonicalAuthorization(mutated)
		mutatedSignature := signature
		mutatedSignature.AuthorizationSHA256 = SHA256(mutatedBody)
		_, err := VerifyWhole(
			manifestBody, fixture(t, "unirouter-ack.json"), fixture(t, "xinsight-ack.json"),
			mutatedBody, mustCanonicalSignature(t, mutatedSignature), trustBody,
		)
		if err == nil || !strings.Contains(err.Error(), "acknowledgement digest mismatch") {
			t.Fatalf("expected acknowledgement binding rejection, got %v", err)
		}
	})
}

func mustCanonicalSignature(t *testing.T, value AuthorizationSignature) []byte {
	t.Helper()
	body, err := CanonicalSignature(value)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestCanonicalAndStrictJSONFailures(t *testing.T) {
	if _, err := DecodeCanonicalManifest(append(fixture(t, "manifest.json"), '\n')); err == nil {
		t.Fatal("non-canonical trailing newline was accepted")
	}
	duplicate := []byte("{\"schema_version\":\"" + ManifestSchemaVersion + "\",\"schema_version\":\"" + ManifestSchemaVersion + "\"}")
	if _, err := DecodeManifest(duplicate); err == nil || !strings.Contains(err.Error(), "duplicate key") {
		t.Fatalf("expected duplicate-key rejection, got %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(fixture(t, "authorization.json"), &document); err != nil {
		t.Fatal(err)
	}
	document["raw_confirmation"] = "must never be accepted"
	body, _ := json.Marshal(document)
	if _, err := DecodeAuthorization(body); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown-field rejection, got %v", err)
	}
}

func TestAuthorizationWindowEndIsExclusive(t *testing.T) {
	authorization, err := DecodeCanonicalAuthorization(fixture(t, "authorization.json"))
	if err != nil {
		t.Fatal(err)
	}
	authorization.ConfirmedAt = authorization.AuthorizationWindow.NotAfter
	if err := ValidateAuthorization(authorization); err == nil {
		t.Fatal("authorization confirmed exactly at not_after was accepted")
	}
}

func TestCommittedInvalidFixturesFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		verify func([]byte) error
	}{
		{"authorization-trust-set-wrong-purpose.json", func(body []byte) error {
			_, err := DecodeTrustSet(body)
			return err
		}},
		{"authorization-signature-source-key-ref.json", func(body []byte) error {
			_, err := DecodeSignature(body)
			return err
		}},
		{"component-ack-rejected.json", func(body []byte) error {
			_, err := DecodeAcknowledgement(body)
			return err
		}},
		{"duplicate-schema-key.json", func(body []byte) error {
			_, err := DecodeManifest(body)
			return err
		}},
		{"manifest-wrong-source-identity.json", func(body []byte) error {
			_, err := DecodeManifest(body)
			return err
		}},
		{"manifest-wrong-request-set-identity.json", func(body []byte) error {
			_, err := DecodeManifest(body)
			return err
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.verify(invalidFixture(t, testCase.name)); err == nil {
				t.Fatal("invalid fixture was accepted")
			}
		})
	}
}

func TestCommittedWrongProfileIdentityFixturesFailForTheNamedIdentity(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"manifest-wrong-source-identity.json", "Source binding contract identity is invalid"},
		{"manifest-wrong-request-set-identity.json", "request-set binding is invalid"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := DecodeManifest(invalidFixture(t, testCase.name))
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("expected %q rejection, got %v", testCase.want, err)
			}
		})
	}
}

func TestPreauthorizationArtifactsContainNoFutureResultBinding(t *testing.T) {
	for _, name := range []string{"manifest.json", "authorization.json"} {
		body := fixture(t, name)
		for _, forbidden := range []string{"package_sha256", "import_receipt_sha256", "terminal_result", "result_sha256"} {
			if bytes.Contains(body, []byte(forbidden)) {
				t.Fatalf("%s contains future result field %q", name, forbidden)
			}
		}
	}
}

func TestReleaseContractRequiresReproducibilityAndZeroEffects(t *testing.T) {
	release := ReleaseManifest{
		SchemaVersion: ReleaseSchemaVersion, Canonicalization: CanonicalizationVersion,
		Version: "0.1.0-rc.1", SourceCommit: strings.Repeat("a", 40), SourceTree: strings.Repeat("b", 40),
		GoVersion: "go1.25.0", Toolchain: "go1.25.12",
		ContractsSHA256:       "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		ValidFixturesSHA256:   "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		InvalidFixturesSHA256: "sha256:3333333333333333333333333333333333333333333333333333333333333333",
		Artifacts: []ReleaseArtifact{{
			Path: "dist/foigent-integration-control-darwin-arm64", OS: "darwin", Architecture: "arm64",
			SizeBytes: 1, SHA256: "sha256:4444444444444444444444444444444444444444444444444444444444444444",
		}},
		DoubleBuildEqual: true, CapabilityBoundary: CapabilityBoundaryOfflineOnly,
		Gates:   []ReleaseGate{{Name: "full_n1", Status: ReleaseGatePass}},
		Effects: ReleaseEffects{},
	}
	body, err := CanonicalRelease(release)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCanonicalRelease(body); err != nil {
		t.Fatal(err)
	}
	release.Effects.WorkerOrTraffic = 1
	if err := ValidateRelease(release); err == nil {
		t.Fatal("release with external traffic effect was accepted")
	}
}
