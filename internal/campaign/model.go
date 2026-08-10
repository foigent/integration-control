package campaign

const (
	CanonicalizationVersion = "foigent.integration-control.canonical-json.v1"

	ManifestSchemaVersion      = "foigent.integration-campaign-manifest.v1"
	AcknowledgementSchema      = "foigent.integration-campaign-component-ack.v1"
	AuthorizationSchemaVersion = "foigent.integration-campaign-authorization.v1"
	SignatureSchemaVersion     = "foigent.integration-campaign-authorization-signature.v1"
	TrustSetSchemaVersion      = "foigent.integration-campaign-authorization-trust-set.v1"
	ReleaseSchemaVersion       = "foigent.integration-control.release.v1"

	ManifestProfile = "poc_source_evidence"

	SourceBindingSchemaVersion   = "unirouter.source-evidence.source-scope.v1"
	RequestSetSchemaVersion      = "unirouter.source-evidence.request-set.v1"
	UniRouterComponentContractID = "unirouter.source-evidence.component-package.v1"
	XInsightComponentContractID  = "xinsight.signed-campaign-package-consumer.v1"

	SourceEvidencePurpose = "source_evidence"
	SourceEvidenceDomain  = "unirouter.source-evidence.component.v1"

	AuthorizationPurpose = "integration_campaign_authorization"
	AuthorizationDomain  = "foigent.integration-campaign-authorization.v1"
	SignatureAlgorithm   = "ed25519"

	PackageClaimScope                  = "unirouter.poc-source-facts.v1"
	PackageScopeState                  = "complete"
	NestedExecutionEvidenceV2State     = "partial"
	ComponentAcknowledgementAccepted   = "accepted"
	AuthorizationTrustKeyActive        = "active"
	AuthorizationTrustKeyDisabled      = "disabled"
	CapabilityBoundaryOfflineOnly      = "offline_files_only"
	ReleaseGatePass                    = "pass"
	ReleaseEffectsExpectedZero         = 0
	MaximumAuthorizationTrustSetKeys   = 16
	MaximumCampaignIdentifierRunes     = 128
	MaximumDecisionIdentifierRunes     = 160
	MaximumOperationalReferenceRunes   = 160
	MaximumContractIdentifierRunes     = 200
	MaximumArtifactSchemaVersionRunes  = 200
	MaximumReleaseArtifactPathRunes    = 240
	MaximumAuthorizationIdentifierRune = 160
)

var RequiredRequestedActions = []string{
	"poc_traffic",
	"unirouter_source_evidence_capture",
	"xinsight_source_evidence_import",
}

var RequiredExplicitExclusions = []string{
	"bootstrap",
	"credential_mutation",
	"data_reset",
	"deployment",
	"production",
	"target_trust_mutation",
	"traffic_outside_frozen_request_set",
	"worker_activation",
}

type ArtifactReference struct {
	SchemaVersion string `json:"schema_version"`
	SHA256        string `json:"sha256"`
}

type ContractReference struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

type AuthorizationWindow struct {
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

type RequestSetReference struct {
	SchemaVersion string `json:"schema_version"`
	SHA256        string `json:"sha256"`
	MaxRequests   int    `json:"max_requests"`
	MaxFacts      int    `json:"max_facts"`
}

type ComponentInput struct {
	Product      string            `json:"product"`
	Role         string            `json:"role"`
	Contract     ContractReference `json:"contract"`
	Release      ArtifactReference `json:"release"`
	ScopeReceipt ArtifactReference `json:"scope_receipt"`
}

type SourceEvidenceSigning struct {
	Purpose                   string `json:"purpose"`
	Domain                    string `json:"domain"`
	SigningKeyRef             string `json:"signing_key_ref"`
	SignerPublicKeyID         string `json:"signer_public_key_id"`
	TrustSetSHA256            string `json:"trust_set_sha256"`
	ClaimScope                string `json:"claim_scope"`
	PackageScopeState         string `json:"package_scope_state"`
	NestedExecutionEvidenceV2 string `json:"nested_execution_evidence_v2"`
	RuntimeCloseMarker        bool   `json:"runtime_close_marker_available"`
}

type AuthorizationSigning struct {
	Purpose        string `json:"purpose"`
	Domain         string `json:"domain"`
	TrustSetSHA256 string `json:"trust_set_sha256"`
}

type Manifest struct {
	SchemaVersion         string                `json:"schema_version"`
	Canonicalization      string                `json:"canonicalization"`
	Profile               string                `json:"profile"`
	DecisionID            string                `json:"decision_id"`
	DecisionSHA256        string                `json:"decision_sha256"`
	CampaignID            string                `json:"campaign_id"`
	FrozenAt              string                `json:"frozen_at"`
	RequestedActions      []string              `json:"requested_actions"`
	RequestedWindow       AuthorizationWindow   `json:"requested_window"`
	ExplicitExclusions    []string              `json:"explicit_exclusions"`
	SourceBinding         ArtifactReference     `json:"source_binding"`
	RequestSet            RequestSetReference   `json:"request_set"`
	Components            []ComponentInput      `json:"components"`
	SourceEvidenceSigning SourceEvidenceSigning `json:"source_evidence_signing"`
	AuthorizationSigning  AuthorizationSigning  `json:"authorization_signing"`
}

type ComponentAcknowledgement struct {
	SchemaVersion              string            `json:"schema_version"`
	Product                    string            `json:"product"`
	CampaignManifestSHA256     string            `json:"campaign_manifest_sha256"`
	ComponentContract          ContractReference `json:"component_contract"`
	ComponentScopeReceipt      ArtifactReference `json:"component_scope_receipt"`
	ProductVerificationReceipt ArtifactReference `json:"product_verification_receipt"`
	VerifierRelease            ContractReference `json:"verifier_release"`
	Status                     string            `json:"status"`
	VerifiedAt                 string            `json:"verified_at"`
}

type ComponentAcknowledgementReference struct {
	Product                      string `json:"product"`
	AcknowledgementSchemaVersion string `json:"acknowledgement_schema_version"`
	AcknowledgementSHA256        string `json:"acknowledgement_sha256"`
	CampaignManifestSHA256       string `json:"campaign_manifest_sha256"`
}

type Authorization struct {
	SchemaVersion             string                              `json:"schema_version"`
	Canonicalization          string                              `json:"canonicalization"`
	AuthorizationID           string                              `json:"authorization_id"`
	DecisionID                string                              `json:"decision_id"`
	CampaignID                string                              `json:"campaign_id"`
	CampaignManifestSHA256    string                              `json:"campaign_manifest_sha256"`
	ComponentAcknowledgements []ComponentAcknowledgementReference `json:"component_acknowledgements"`
	AuthorizedActions         []string                            `json:"authorized_actions"`
	AuthorizationWindow       AuthorizationWindow                 `json:"authorization_window"`
	ConfirmedAt               string                              `json:"confirmed_at"`
	ConfirmationSHA256        string                              `json:"confirmation_sha256"`
}

type AuthorizationSignature struct {
	SchemaVersion       string `json:"schema_version"`
	Purpose             string `json:"purpose"`
	Domain              string `json:"domain"`
	SigningKeyRef       string `json:"signing_key_ref"`
	SignerPublicKeyID   string `json:"signer_public_key_id"`
	TrustSetSHA256      string `json:"trust_set_sha256"`
	AuthorizationSHA256 string `json:"authorization_sha256"`
	Algorithm           string `json:"algorithm"`
	Value               string `json:"value"`
}

type AuthorizationTrustSet struct {
	SchemaVersion string                  `json:"schema_version"`
	Purpose       string                  `json:"purpose"`
	Domain        string                  `json:"domain"`
	TrustSetRef   string                  `json:"trust_set_ref"`
	Keys          []AuthorizationTrustKey `json:"keys"`
}

type AuthorizationTrustKey struct {
	SigningKeyRef     string `json:"signing_key_ref"`
	SignerPublicKeyID string `json:"signer_public_key_id"`
	Algorithm         string `json:"algorithm"`
	PublicKey         string `json:"public_key"`
	Status            string `json:"status"`
	ValidFrom         string `json:"valid_from"`
	ValidUntil        string `json:"valid_until"`
}

type ReleaseArtifact struct {
	Path         string `json:"path"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	SizeBytes    int64  `json:"size_bytes"`
	SHA256       string `json:"sha256"`
}

type ReleaseGate struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ReleaseEffects struct {
	TargetAccess          int `json:"target_access"`
	ProductRepositoryEdit int `json:"product_repository_edit"`
	ProductDBOrAPIAccess  int `json:"product_db_or_api_access"`
	CredentialOrKeyAccess int `json:"credential_or_key_access"`
	WorkerOrTraffic       int `json:"worker_or_traffic"`
}

type ReleaseManifest struct {
	SchemaVersion         string            `json:"schema_version"`
	Canonicalization      string            `json:"canonicalization"`
	Version               string            `json:"version"`
	SourceCommit          string            `json:"source_commit"`
	SourceTree            string            `json:"source_tree"`
	GoVersion             string            `json:"go_version"`
	Toolchain             string            `json:"toolchain"`
	ContractsSHA256       string            `json:"contracts_sha256"`
	ValidFixturesSHA256   string            `json:"valid_fixtures_sha256"`
	InvalidFixturesSHA256 string            `json:"invalid_fixtures_sha256"`
	Artifacts             []ReleaseArtifact `json:"artifacts"`
	DoubleBuildEqual      bool              `json:"double_build_equal"`
	CapabilityBoundary    string            `json:"capability_boundary"`
	Gates                 []ReleaseGate     `json:"gates"`
	Effects               ReleaseEffects    `json:"effects"`
}
