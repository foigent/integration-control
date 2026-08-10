package campaign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const JSONSchemaDialect = "https://json-schema.org/draft/2020-12/schema"

const (
	ManifestSchemaFile      = "integration-campaign-manifest.v1.schema.json"
	AuthorizationSchemaFile = "integration-campaign-authorization.v1.schema.json"
	SignatureSchemaFile     = "integration-campaign-authorization-signature.v1.schema.json"
	TrustSetSchemaFile      = "integration-campaign-authorization-trust-set.v1.schema.json"
	ReleaseSchemaFile       = "integration-control-release.v1.schema.json"
)

type schemaObject map[string]any

func ProjectedSchemas() (map[string][]byte, error) {
	documents := map[string]schemaObject{
		ManifestSchemaFile:      manifestSchema(),
		AuthorizationSchemaFile: authorizationSchema(),
		SignatureSchemaFile:     signatureSchema(),
		TrustSetSchemaFile:      trustSetSchema(),
		ReleaseSchemaFile:       releaseSchema(),
	}
	projected := make(map[string][]byte, len(documents))
	for name, document := range documents {
		body, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("project schema %s: %w", name, err)
		}
		projected[name] = append(body, '\n')
	}
	return projected, nil
}

func SchemaFiles() []string {
	files := []string{ManifestSchemaFile, AuthorizationSchemaFile, SignatureSchemaFile, TrustSetSchemaFile, ReleaseSchemaFile}
	sort.Strings(files)
	return files
}

func WriteProjectedSchemas(directory string) error {
	projected, err := ProjectedSchemas()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create contract directory: %w", err)
	}
	for _, name := range SchemaFiles() {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, projected[name], 0o644); err != nil {
			return fmt.Errorf("write schema %s: %w", name, err)
		}
	}
	return nil
}

func CheckProjectedSchemas(directory string) error {
	projected, err := ProjectedSchemas()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read contract directory: %w", err)
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		expected, known := projected[entry.Name()]
		if !known {
			return fmt.Errorf("unexpected projected schema file %q", entry.Name())
		}
		body, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(body, expected) {
			return fmt.Errorf("projected schema drift detected in %q", entry.Name())
		}
		seen[entry.Name()] = true
	}
	for _, name := range SchemaFiles() {
		if !seen[name] {
			return fmt.Errorf("projected schema %q is missing", name)
		}
	}
	return nil
}

func manifestSchema() schemaObject {
	component := func(product, role string) schemaObject {
		return objectSchema(schemaObject{
			"product":       constString(product),
			"role":          constString(role),
			"contract":      ref("#/$defs/contract_reference"),
			"release":       ref("#/$defs/artifact_reference"),
			"scope_receipt": ref("#/$defs/artifact_reference"),
		}, "product", "role", "contract", "release", "scope_receipt")
	}
	return rootSchema(
		"Product-neutral immutable preauthorization scope for one POC Source-evidence campaign.",
		ManifestSchemaVersion,
		objectSchema(schemaObject{
			"schema_version":      constString(ManifestSchemaVersion),
			"canonicalization":    constString(CanonicalizationVersion),
			"profile":             constString(ManifestProfile),
			"decision_id":         portableString(MaximumDecisionIdentifierRunes),
			"decision_sha256":     digestString(),
			"campaign_id":         portableString(MaximumCampaignIdentifierRunes),
			"frozen_at":           timestampString(),
			"requested_actions":   exactStringArray(RequiredRequestedActions),
			"requested_window":    ref("#/$defs/authorization_window"),
			"explicit_exclusions": exactStringArray(RequiredExplicitExclusions),
			"source_binding":      ref("#/$defs/artifact_reference"),
			"request_set": objectSchema(schemaObject{
				"schema_version": portableString(MaximumArtifactSchemaVersionRunes),
				"sha256":         digestString(),
				"max_requests":   schemaObject{"const": 20},
				"max_facts":      schemaObject{"const": 200},
			}, "schema_version", "sha256", "max_requests", "max_facts"),
			"components": schemaObject{
				"type":        "array",
				"minItems":    2,
				"maxItems":    2,
				"prefixItems": []any{component("unirouter", "source_evidence_producer"), component("xinsight", "source_evidence_consumer")},
			},
			"source_evidence_signing": objectSchema(schemaObject{
				"purpose":                        constString(SourceEvidencePurpose),
				"domain":                         constString(SourceEvidenceDomain),
				"signing_key_ref":                prefixedPortableString("^source-evidence-"),
				"signer_public_key_id":           publicKeyIDString(),
				"trust_set_sha256":               digestString(),
				"claim_scope":                    constString(PackageClaimScope),
				"package_scope_state":            constString(PackageScopeState),
				"nested_execution_evidence_v2":   constString(NestedExecutionEvidenceV2State),
				"runtime_close_marker_available": schemaObject{"const": false},
			}, "purpose", "domain", "signing_key_ref", "signer_public_key_id", "trust_set_sha256", "claim_scope", "package_scope_state", "nested_execution_evidence_v2", "runtime_close_marker_available"),
			"authorization_signing": objectSchema(schemaObject{
				"purpose":          constString(AuthorizationPurpose),
				"domain":           constString(AuthorizationDomain),
				"trust_set_sha256": digestString(),
			}, "purpose", "domain", "trust_set_sha256"),
		}, "schema_version", "canonicalization", "profile", "decision_id", "decision_sha256", "campaign_id", "frozen_at", "requested_actions", "requested_window", "explicit_exclusions", "source_binding", "request_set", "components", "source_evidence_signing", "authorization_signing"),
		schemaObject{
			"artifact_reference":   artifactReferenceSchema(),
			"contract_reference":   contractReferenceSchema(),
			"authorization_window": authorizationWindowSchema(),
		},
	)
}

func authorizationSchema() schemaObject {
	ackReference := func(product string) schemaObject {
		return objectSchema(schemaObject{
			"product":                        constString(product),
			"acknowledgement_schema_version": constString(AcknowledgementSchema),
			"acknowledgement_sha256":         digestString(),
			"campaign_manifest_sha256":       digestString(),
		}, "product", "acknowledgement_schema_version", "acknowledgement_sha256", "campaign_manifest_sha256")
	}
	genericAckReference := objectSchema(schemaObject{
		"product":                        schemaObject{"enum": []string{"unirouter", "xinsight"}},
		"acknowledgement_schema_version": constString(AcknowledgementSchema),
		"acknowledgement_sha256":         digestString(),
		"campaign_manifest_sha256":       digestString(),
	}, "product", "acknowledgement_schema_version", "acknowledgement_sha256", "campaign_manifest_sha256")
	ackEnvelope := objectSchema(schemaObject{
		"schema_version":               constString(AcknowledgementSchema),
		"product":                      schemaObject{"enum": []string{"unirouter", "xinsight"}},
		"campaign_manifest_sha256":     digestString(),
		"component_contract":           ref("#/$defs/contract_reference"),
		"component_scope_receipt":      ref("#/$defs/artifact_reference"),
		"product_verification_receipt": ref("#/$defs/artifact_reference"),
		"verifier_release":             ref("#/$defs/contract_reference"),
		"status":                       constString(ComponentAcknowledgementAccepted),
		"verified_at":                  timestampString(),
	}, "schema_version", "product", "campaign_manifest_sha256", "component_contract", "component_scope_receipt", "product_verification_receipt", "verifier_release", "status", "verified_at")
	return rootSchema(
		"User decision binding for one exact manifest and two product-owned component acknowledgements.",
		AuthorizationSchemaVersion,
		objectSchema(schemaObject{
			"schema_version":           constString(AuthorizationSchemaVersion),
			"canonicalization":         constString(CanonicalizationVersion),
			"authorization_id":         portableString(MaximumAuthorizationIdentifierRune),
			"decision_id":              portableString(MaximumDecisionIdentifierRunes),
			"campaign_id":              portableString(MaximumCampaignIdentifierRunes),
			"campaign_manifest_sha256": digestString(),
			"component_acknowledgements": schemaObject{
				"type":        "array",
				"minItems":    2,
				"maxItems":    2,
				"prefixItems": []any{ackReference("unirouter"), ackReference("xinsight")},
			},
			"authorized_actions":   exactStringArray(RequiredRequestedActions),
			"authorization_window": ref("#/$defs/authorization_window"),
			"confirmed_at":         timestampString(),
			"confirmation_sha256":  digestString(),
		}, "schema_version", "canonicalization", "authorization_id", "decision_id", "campaign_id", "campaign_manifest_sha256", "component_acknowledgements", "authorized_actions", "authorization_window", "confirmed_at", "confirmation_sha256"),
		schemaObject{
			"acknowledgement_reference": genericAckReference,
			"component_acknowledgement": ackEnvelope,
			"artifact_reference":        artifactReferenceSchema(),
			"contract_reference":        contractReferenceSchema(),
			"authorization_window":      authorizationWindowSchema(),
		},
	)
}

func signatureSchema() schemaObject {
	return rootSchema(
		"Detached purpose-separated Ed25519 signature envelope for a canonical authorization artifact.",
		SignatureSchemaVersion,
		objectSchema(schemaObject{
			"schema_version":       constString(SignatureSchemaVersion),
			"purpose":              constString(AuthorizationPurpose),
			"domain":               constString(AuthorizationDomain),
			"signing_key_ref":      prefixedPortableString("^integration-authorization-"),
			"signer_public_key_id": publicKeyIDString(),
			"trust_set_sha256":     digestString(),
			"authorization_sha256": digestString(),
			"algorithm":            constString(SignatureAlgorithm),
			"value": schemaObject{
				"type":      "string",
				"pattern":   "^[A-Za-z0-9+/]{86}==$",
				"maxLength": 88,
			},
		}, "schema_version", "purpose", "domain", "signing_key_ref", "signer_public_key_id", "trust_set_sha256", "authorization_sha256", "algorithm", "value"),
		nil,
	)
}

func trustSetSchema() schemaObject {
	key := objectSchema(schemaObject{
		"signing_key_ref":      prefixedPortableString("^integration-authorization-"),
		"signer_public_key_id": publicKeyIDString(),
		"algorithm":            constString(SignatureAlgorithm),
		"public_key": schemaObject{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9+/]{43}=$",
			"maxLength": 44,
		},
		"status":      schemaObject{"enum": []string{AuthorizationTrustKeyActive, AuthorizationTrustKeyDisabled}},
		"valid_from":  timestampString(),
		"valid_until": timestampString(),
	}, "signing_key_ref", "signer_public_key_id", "algorithm", "public_key", "status", "valid_from", "valid_until")
	return rootSchema(
		"Immutable purpose-specific public authorization trust set.",
		TrustSetSchemaVersion,
		objectSchema(schemaObject{
			"schema_version": constString(TrustSetSchemaVersion),
			"purpose":        constString(AuthorizationPurpose),
			"domain":         constString(AuthorizationDomain),
			"trust_set_ref":  prefixedPortableString("^integration-authorization-trust-"),
			"keys": schemaObject{
				"type":     "array",
				"minItems": 1,
				"maxItems": MaximumAuthorizationTrustSetKeys,
				"items":    key,
			},
		}, "schema_version", "purpose", "domain", "trust_set_ref", "keys"),
		nil,
	)
}

func releaseSchema() schemaObject {
	artifact := objectSchema(schemaObject{
		"path":         portableString(MaximumReleaseArtifactPathRunes),
		"os":           portableString(40),
		"architecture": portableString(40),
		"size_bytes":   schemaObject{"type": "integer", "minimum": 1},
		"sha256":       digestString(),
	}, "path", "os", "architecture", "size_bytes", "sha256")
	gate := objectSchema(schemaObject{
		"name":   portableString(120),
		"status": constString(ReleaseGatePass),
	}, "name", "status")
	effects := objectSchema(schemaObject{
		"target_access":            schemaObject{"const": ReleaseEffectsExpectedZero},
		"product_repository_edit":  schemaObject{"const": ReleaseEffectsExpectedZero},
		"product_db_or_api_access": schemaObject{"const": ReleaseEffectsExpectedZero},
		"credential_or_key_access": schemaObject{"const": ReleaseEffectsExpectedZero},
		"worker_or_traffic":        schemaObject{"const": ReleaseEffectsExpectedZero},
	}, "target_access", "product_repository_edit", "product_db_or_api_access", "credential_or_key_access", "worker_or_traffic")
	return rootSchema(
		"Reproducible local release evidence for the offline neutral control asset.",
		ReleaseSchemaVersion,
		objectSchema(schemaObject{
			"schema_version":          constString(ReleaseSchemaVersion),
			"canonicalization":        constString(CanonicalizationVersion),
			"version":                 schemaObject{"type": "string", "pattern": releaseVersion.String()},
			"source_commit":           schemaObject{"type": "string", "pattern": gitObjectPattern.String()},
			"source_tree":             schemaObject{"type": "string", "pattern": gitObjectPattern.String()},
			"go_version":              constString("go1.25.0"),
			"toolchain":               constString("go1.25.12"),
			"contracts_sha256":        digestString(),
			"valid_fixtures_sha256":   digestString(),
			"invalid_fixtures_sha256": digestString(),
			"artifacts": schemaObject{
				"type":     "array",
				"minItems": 1,
				"items":    artifact,
			},
			"double_build_equal":  schemaObject{"const": true},
			"capability_boundary": constString(CapabilityBoundaryOfflineOnly),
			"gates": schemaObject{
				"type":     "array",
				"minItems": 1,
				"items":    gate,
			},
			"effects": effects,
		}, "schema_version", "canonicalization", "version", "source_commit", "source_tree", "go_version", "toolchain", "contracts_sha256", "valid_fixtures_sha256", "invalid_fixtures_sha256", "artifacts", "double_build_equal", "capability_boundary", "gates", "effects"),
		nil,
	)
}

func rootSchema(description, identifier string, body schemaObject, definitions schemaObject) schemaObject {
	root := schemaObject{
		"$schema":     JSONSchemaDialect,
		"$id":         "https://schemas.foigent.local/" + identifier + ".schema.json",
		"title":       identifier,
		"description": description,
	}
	for key, value := range body {
		root[key] = value
	}
	if len(definitions) != 0 {
		root["$defs"] = definitions
	}
	return root
}

func objectSchema(properties schemaObject, required ...string) schemaObject {
	return schemaObject{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
		"required":             required,
	}
}

func artifactReferenceSchema() schemaObject {
	return objectSchema(schemaObject{
		"schema_version": portableString(MaximumArtifactSchemaVersionRunes),
		"sha256":         digestString(),
	}, "schema_version", "sha256")
}

func contractReferenceSchema() schemaObject {
	return objectSchema(schemaObject{
		"id":     portableString(MaximumContractIdentifierRunes),
		"sha256": digestString(),
	}, "id", "sha256")
}

func authorizationWindowSchema() schemaObject {
	return objectSchema(schemaObject{
		"not_before": timestampString(),
		"not_after":  timestampString(),
	}, "not_before", "not_after")
}

func exactStringArray(values []string) schemaObject {
	prefix := make([]any, 0, len(values))
	for _, value := range values {
		prefix = append(prefix, constString(value))
	}
	return schemaObject{
		"type":        "array",
		"minItems":    len(values),
		"maxItems":    len(values),
		"prefixItems": prefix,
	}
}

func ref(path string) schemaObject {
	return schemaObject{"$ref": path}
}

func constString(value string) schemaObject {
	return schemaObject{"const": value, "type": "string"}
}

func digestString() schemaObject {
	return schemaObject{"type": "string", "pattern": digestPattern.String()}
}

func publicKeyIDString() schemaObject {
	return schemaObject{"type": "string", "pattern": publicKeyIDPattern.String()}
}

func timestampString() schemaObject {
	return schemaObject{"type": "string", "format": "date-time"}
}

func portableString(maximum int) schemaObject {
	return schemaObject{"type": "string", "pattern": portableIDPattern.String(), "maxLength": maximum}
}

func prefixedPortableString(prefix string) schemaObject {
	return schemaObject{"type": "string", "pattern": prefix + "[A-Za-z0-9._:/-]+$", "maxLength": MaximumOperationalReferenceRunes}
}

func VerifyProjectedSchemaDocument(body []byte) error {
	if len(body) == 0 || !json.Valid(body) {
		return errors.New("projected JSON Schema is invalid JSON")
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		return err
	}
	if document["$schema"] != JSONSchemaDialect || document["type"] != "object" {
		return errors.New("projected JSON Schema root identity is invalid")
	}
	return nil
}
