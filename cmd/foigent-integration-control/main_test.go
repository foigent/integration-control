package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/foigent/integration-control/internal/campaign"
)

func validPath(name string) string {
	return filepath.Join("..", "..", "fixtures", "v1", "valid", name)
}

func TestAuthorizationVerifyCommand(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{
		"authorization", "verify",
		"--manifest", validPath("manifest.json"),
		"--unirouter-ack", validPath("unirouter-ack.json"),
		"--xinsight-ack", validPath("xinsight-ack.json"),
		"--authorization", validPath("authorization.json"),
		"--signature", validPath("authorization-signature.json"),
		"--trust-set", validPath("authorization-trust-set.json"),
	}, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\"valid\":true") ||
		!strings.Contains(output.String(), "\"capability_boundary\":\"offline_files_only\"") {
		t.Fatalf("unexpected verification output: %s", output.String())
	}
}

func TestSchemaCheckCommand(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{"schemas", "check", "--dir", filepath.Join("..", "..", "contracts", "v1")}, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\"schema_count\":5") {
		t.Fatalf("unexpected schema check output: %s", output.String())
	}
}

func TestManifestBuildCommandIsReproducible(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{
		"manifest", "build",
		"--draft", validPath("manifest-draft.json"),
		"--decision", validPath("decision.txt"),
		"--source-binding", validPath("source-binding.json"),
		"--request-set", validPath("request-set.json"),
		"--source-evidence-trust-set", validPath("source-evidence-trust-set.json"),
		"--authorization-trust-set", validPath("authorization-trust-set.json"),
		"--unirouter-contract", validPath("unirouter-contract.json"),
		"--unirouter-release", validPath("unirouter-release.json"),
		"--unirouter-scope", validPath("unirouter-scope.json"),
		"--xinsight-contract", validPath("xinsight-contract.json"),
		"--xinsight-release", validPath("xinsight-release.json"),
		"--xinsight-scope", validPath("xinsight-scope.json"),
		"--out", "-",
	}, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(validPath("manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output.Bytes(), expected) {
		t.Fatal("CLI manifest build is not byte-reproducible")
	}
}

func TestAuthorizationBuildHashesButDoesNotEmitConfirmation(t *testing.T) {
	confirmation := []byte("synthetic CLI test input, not a user confirmation")
	confirmationPath := filepath.Join(t.TempDir(), "confirmation.txt")
	if err := os.WriteFile(confirmationPath, confirmation, 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := run([]string{
		"authorization", "build",
		"--draft", validPath("authorization-draft.json"),
		"--manifest", validPath("manifest.json"),
		"--unirouter-ack", validPath("unirouter-ack.json"),
		"--xinsight-ack", validPath("xinsight-ack.json"),
		"--confirmation", confirmationPath,
		"--out", "-",
	}, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(output.Bytes(), confirmation) {
		t.Fatal("CLI emitted raw confirmation material")
	}
	authorization, err := campaign.DecodeCanonicalAuthorization(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if authorization.ConfirmationSHA256 != campaign.SHA256(confirmation) {
		t.Fatal("CLI did not bind exact confirmation bytes")
	}
}

func TestPreimageCommandIsDomainSeparated(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"authorization", "preimage", "--authorization", validPath("authorization.json"), "--out", "-"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(output.Bytes(), []byte(campaign.AuthorizationPreimagePrefix)) {
		t.Fatal("authorization preimage is not domain and purpose separated")
	}
}

func TestOutputNeverOverwritesExistingFile(t *testing.T) {
	existing := filepath.Join(t.TempDir(), "existing.json")
	if err := os.WriteFile(existing, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := run([]string{
		"manifest", "build",
		"--draft", validPath("manifest-draft.json"),
		"--decision", validPath("decision.txt"),
		"--source-binding", validPath("source-binding.json"),
		"--request-set", validPath("request-set.json"),
		"--source-evidence-trust-set", validPath("source-evidence-trust-set.json"),
		"--authorization-trust-set", validPath("authorization-trust-set.json"),
		"--unirouter-contract", validPath("unirouter-contract.json"),
		"--unirouter-release", validPath("unirouter-release.json"),
		"--unirouter-scope", validPath("unirouter-scope.json"),
		"--xinsight-contract", validPath("xinsight-contract.json"),
		"--xinsight-release", validPath("xinsight-release.json"),
		"--xinsight-scope", validPath("xinsight-scope.json"),
		"--out", existing,
	}, &output, &output)
	if err == nil || !strings.Contains(err.Error(), "without overwrite") {
		t.Fatalf("expected no-overwrite failure, got %v", err)
	}
	body, readErr := os.ReadFile(existing)
	if readErr != nil || string(body) != "preserve" {
		t.Fatal("existing output was modified")
	}
}

func TestOutputPublishesCompleteCreateOnlyArtifact(t *testing.T) {
	directory := t.TempDir()
	outputPath := filepath.Join(directory, "manifest.json")
	var output bytes.Buffer
	err := run([]string{
		"manifest", "build",
		"--draft", validPath("manifest-draft.json"),
		"--decision", validPath("decision.txt"),
		"--source-binding", validPath("source-binding.json"),
		"--request-set", validPath("request-set.json"),
		"--source-evidence-trust-set", validPath("source-evidence-trust-set.json"),
		"--authorization-trust-set", validPath("authorization-trust-set.json"),
		"--unirouter-contract", validPath("unirouter-contract.json"),
		"--unirouter-release", validPath("unirouter-release.json"),
		"--unirouter-scope", validPath("unirouter-scope.json"),
		"--xinsight-contract", validPath("xinsight-contract.json"),
		"--xinsight-release", validPath("xinsight-release.json"),
		"--xinsight-scope", validPath("xinsight-scope.json"),
		"--out", outputPath,
	}, &output, &output)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(validPath("manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, expected) {
		t.Fatal("published output is incomplete or changed")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "manifest.json" {
		t.Fatalf("temporary output was not cleaned: %#v", entries)
	}
}
