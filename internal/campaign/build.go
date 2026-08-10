package campaign

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const maximumConfirmationBytes = 64 << 10

// ManifestBuildInputs are exact, product-owned or coordinator-owned immutable
// documents. The neutral builder only validates their declared schema identities
// and hashes their bytes; it does not interpret or assert their product facts.
type ManifestBuildInputs struct {
	Decision               []byte
	SourceBinding          []byte
	RequestSet             []byte
	SourceEvidenceTrustSet []byte
	AuthorizationTrustSet  []byte
	UniRouterContract      []byte
	UniRouterRelease       []byte
	UniRouterScope         []byte
	XInsightContract       []byte
	XInsightRelease        []byte
	XInsightScope          []byte
}

// BuildManifestFromDraft replaces every content digest with a digest computed
// from the exact input bytes. A draft containing any pre-filled content digest
// is rejected so an operator-entered digest cannot become authority.
func BuildManifestFromDraft(draftBody []byte, inputs ManifestBuildInputs) (Manifest, []byte, error) {
	draft, err := decodeStrict[Manifest](draftBody)
	if err != nil {
		return Manifest{}, nil, err
	}
	if err := requireBlankManifestDigests(draft); err != nil {
		return Manifest{}, nil, err
	}
	if len(inputs.Decision) == 0 {
		return Manifest{}, nil, errors.New("decision input is empty")
	}
	if err := bindArtifact(&draft.SourceBinding, inputs.SourceBinding); err != nil {
		return Manifest{}, nil, fmt.Errorf("bind Source: %w", err)
	}
	if err := bindArtifact(&ArtifactReference{SchemaVersion: draft.RequestSet.SchemaVersion}, inputs.RequestSet); err != nil {
		return Manifest{}, nil, fmt.Errorf("bind request set: %w", err)
	}
	draft.RequestSet.SHA256 = SHA256(inputs.RequestSet)
	if _, err := documentSchemaVersion(inputs.SourceEvidenceTrustSet); err != nil {
		return Manifest{}, nil, fmt.Errorf("bind Source-evidence trust set: %w", err)
	}
	if _, err := DecodeCanonicalTrustSet(inputs.AuthorizationTrustSet); err != nil {
		return Manifest{}, nil, fmt.Errorf("bind authorization trust set: %w", err)
	}
	draft.SourceEvidenceSigning.TrustSetSHA256 = SHA256(inputs.SourceEvidenceTrustSet)
	draft.AuthorizationSigning.TrustSetSHA256 = SHA256(inputs.AuthorizationTrustSet)

	material := []struct {
		name     string
		contract []byte
		release  []byte
		scope    []byte
	}{
		{name: "unirouter", contract: inputs.UniRouterContract, release: inputs.UniRouterRelease, scope: inputs.UniRouterScope},
		{name: "xinsight", contract: inputs.XInsightContract, release: inputs.XInsightRelease, scope: inputs.XInsightScope},
	}
	for index := range material {
		if len(material[index].contract) == 0 {
			return Manifest{}, nil, fmt.Errorf("bind %s contract: input is empty", material[index].name)
		}
		draft.Components[index].Contract.SHA256 = SHA256(material[index].contract)
		if err := bindArtifact(&draft.Components[index].Release, material[index].release); err != nil {
			return Manifest{}, nil, fmt.Errorf("bind %s release: %w", material[index].name, err)
		}
		if err := bindArtifact(&draft.Components[index].ScopeReceipt, material[index].scope); err != nil {
			return Manifest{}, nil, fmt.Errorf("bind %s scope receipt: %w", material[index].name, err)
		}
	}
	draft.DecisionSHA256 = SHA256(inputs.Decision)
	canonical, err := CanonicalManifest(draft)
	if err != nil {
		return Manifest{}, nil, err
	}
	return draft, canonical, nil
}

func requireBlankManifestDigests(value Manifest) error {
	if value.DecisionSHA256 != "" || value.SourceBinding.SHA256 != "" || value.RequestSet.SHA256 != "" ||
		value.SourceEvidenceSigning.TrustSetSHA256 != "" || value.AuthorizationSigning.TrustSetSHA256 != "" {
		return errors.New("manifest draft contains a pre-filled coordinator input digest")
	}
	if len(value.Components) != 2 {
		return errors.New("manifest draft must declare exactly two product components")
	}
	for _, component := range value.Components {
		if component.Contract.SHA256 != "" || component.Release.SHA256 != "" || component.ScopeReceipt.SHA256 != "" {
			return errors.New("manifest draft contains a pre-filled product input digest")
		}
	}
	return nil
}

func bindArtifact(reference *ArtifactReference, body []byte) error {
	if len(body) == 0 {
		return errors.New("input is empty")
	}
	schemaVersion, err := documentSchemaVersion(body)
	if err != nil {
		return err
	}
	if schemaVersion != reference.SchemaVersion {
		return fmt.Errorf("schema identity mismatch: expected %q, got %q", reference.SchemaVersion, schemaVersion)
	}
	reference.SHA256 = SHA256(body)
	return nil
}

// BuildAuthorizationFromDraft binds exact canonical acknowledgement bytes and
// the hash of the confirmation text without copying that text into the output.
func BuildAuthorizationFromDraft(draftBody, manifestBody, uniRouterAckBody, xInsightAckBody, confirmation []byte) (Authorization, []byte, error) {
	draft, err := decodeStrict[Authorization](draftBody)
	if err != nil {
		return Authorization{}, nil, err
	}
	if draft.CampaignManifestSHA256 != "" || draft.ConfirmationSHA256 != "" || len(draft.ComponentAcknowledgements) != 0 {
		return Authorization{}, nil, errors.New("authorization draft contains pre-filled derived bindings")
	}
	if len(confirmation) == 0 || len(confirmation) > maximumConfirmationBytes || !utf8.Valid(confirmation) {
		return Authorization{}, nil, errors.New("confirmation input size or UTF-8 encoding is invalid")
	}
	manifest, err := DecodeCanonicalManifest(manifestBody)
	if err != nil {
		return Authorization{}, nil, fmt.Errorf("decode canonical manifest: %w", err)
	}
	manifestSHA := SHA256(manifestBody)
	ackBodies := [][]byte{uniRouterAckBody, xInsightAckBody}
	draft.ComponentAcknowledgements = make([]ComponentAcknowledgementReference, 0, 2)
	for index, body := range ackBodies {
		acknowledgement, ackErr := DecodeCanonicalAcknowledgement(body)
		if ackErr != nil {
			return Authorization{}, nil, fmt.Errorf("decode canonical acknowledgement[%d]: %w", index, ackErr)
		}
		if ackErr = ValidateAcknowledgementAgainstManifest(acknowledgement, manifest); ackErr != nil {
			return Authorization{}, nil, fmt.Errorf("verify acknowledgement[%d]: %w", index, ackErr)
		}
		draft.ComponentAcknowledgements = append(draft.ComponentAcknowledgements, ComponentAcknowledgementReference{
			Product:                      acknowledgement.Product,
			AcknowledgementSchemaVersion: acknowledgement.SchemaVersion,
			AcknowledgementSHA256:        SHA256(body),
			CampaignManifestSHA256:       acknowledgement.CampaignManifestSHA256,
		})
	}
	draft.CampaignManifestSHA256 = manifestSHA
	draft.ConfirmationSHA256 = SHA256(confirmation)
	if err := ValidateAuthorizationAgainstManifest(draft, manifest); err != nil {
		return Authorization{}, nil, err
	}
	if err := validateAcknowledgementChronology(manifest, draft, ackBodies); err != nil {
		return Authorization{}, nil, err
	}
	canonical, err := CanonicalAuthorization(draft)
	if err != nil {
		return Authorization{}, nil, err
	}
	return draft, canonical, nil
}

func validateAcknowledgementChronology(manifest Manifest, authorization Authorization, acknowledgementBodies [][]byte) error {
	frozenAt, _ := parseCanonicalTime("manifest frozen_at", manifest.FrozenAt)
	confirmedAt, _ := parseCanonicalTime("authorization confirmed_at", authorization.ConfirmedAt)
	for index, body := range acknowledgementBodies {
		acknowledgement, err := DecodeCanonicalAcknowledgement(body)
		if err != nil {
			return fmt.Errorf("acknowledgement[%d] chronology input: %w", index, err)
		}
		verifiedAt, _ := parseCanonicalTime("acknowledgement verified_at", acknowledgement.VerifiedAt)
		if verifiedAt.Before(frozenAt) || verifiedAt.After(confirmedAt) {
			return fmt.Errorf("acknowledgement[%d] verification time is outside manifest-freeze to user-confirmation interval", index)
		}
	}
	return nil
}
