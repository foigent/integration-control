package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/foigent/integration-control/internal/campaign"
)

const maximumInputBytes = 4 << 20

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		usage(stderr)
		return errors.New("a command is required")
	}
	switch arguments[0] {
	case "schemas":
		return runSchemas(arguments[1:], stdout, stderr)
	case "manifest":
		return runManifest(arguments[1:], stdout, stderr)
	case "ack":
		return runAcknowledgement(arguments[1:], stdout, stderr)
	case "authorization":
		return runAuthorization(arguments[1:], stdout, stderr)
	case "trust-set":
		return runTrustSet(arguments[1:], stdout, stderr)
	case "release":
		return runRelease(arguments[1:], stdout, stderr)
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func runSchemas(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("schemas requires write or check")
	}
	set := flag.NewFlagSet("schemas "+arguments[0], flag.ContinueOnError)
	set.SetOutput(stderr)
	directory := set.String("dir", "contracts/v1", "projected schema directory")
	if err := set.Parse(arguments[1:]); err != nil {
		return err
	}
	if set.NArg() != 0 {
		return errors.New("schemas command does not accept positional arguments")
	}
	switch arguments[0] {
	case "write":
		if err := campaign.WriteProjectedSchemas(*directory); err != nil {
			return err
		}
	case "check":
		if err := campaign.CheckProjectedSchemas(*directory); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown schemas operation %q", arguments[0])
	}
	return writeReport(stdout, map[string]any{"valid": true, "schema_count": len(campaign.SchemaFiles())})
}

func runManifest(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("manifest requires build or verify")
	}
	switch arguments[0] {
	case "build":
		set := flag.NewFlagSet("manifest build", flag.ContinueOnError)
		set.SetOutput(stderr)
		draft := set.String("draft", "", "manifest draft JSON")
		decision := set.String("decision", "", "exact coordinator decision artifact")
		source := set.String("source-binding", "", "exact typed Source-binding document")
		requestSet := set.String("request-set", "", "exact frozen request-set document")
		sourceTrust := set.String("source-evidence-trust-set", "", "exact product-owned Source-evidence trust set")
		authorizationTrust := set.String("authorization-trust-set", "", "exact canonical neutral authorization trust set")
		uniContract := set.String("unirouter-contract", "", "exact immutable UniRouter component contract")
		uniRelease := set.String("unirouter-release", "", "exact typed UniRouter release identity")
		uniScope := set.String("unirouter-scope", "", "exact typed UniRouter scope receipt")
		xContract := set.String("xinsight-contract", "", "exact immutable X-Insight component contract")
		xRelease := set.String("xinsight-release", "", "exact typed X-Insight release identity")
		xScope := set.String("xinsight-scope", "", "exact typed X-Insight scope receipt")
		output := set.String("out", "-", "new output path, or - for stdout")
		if err := set.Parse(arguments[1:]); err != nil {
			return err
		}
		if set.NArg() != 0 {
			return errors.New("manifest build does not accept positional arguments")
		}
		paths := map[string]string{
			"draft": *draft, "decision": *decision, "source-binding": *source, "request-set": *requestSet,
			"source-evidence-trust-set": *sourceTrust, "authorization-trust-set": *authorizationTrust,
			"unirouter-contract": *uniContract, "unirouter-release": *uniRelease, "unirouter-scope": *uniScope,
			"xinsight-contract": *xContract, "xinsight-release": *xRelease, "xinsight-scope": *xScope,
		}
		bodies, err := readRequiredFiles(paths)
		if err != nil {
			return err
		}
		_, canonical, err := campaign.BuildManifestFromDraft(bodies["draft"], campaign.ManifestBuildInputs{
			Decision: bodies["decision"], SourceBinding: bodies["source-binding"], RequestSet: bodies["request-set"],
			SourceEvidenceTrustSet: bodies["source-evidence-trust-set"], AuthorizationTrustSet: bodies["authorization-trust-set"],
			UniRouterContract: bodies["unirouter-contract"], UniRouterRelease: bodies["unirouter-release"], UniRouterScope: bodies["unirouter-scope"],
			XInsightContract: bodies["xinsight-contract"], XInsightRelease: bodies["xinsight-release"], XInsightScope: bodies["xinsight-scope"],
		})
		if err != nil {
			return err
		}
		return writeExclusive(*output, canonical, stdout)
	case "verify":
		set := flag.NewFlagSet("manifest verify", flag.ContinueOnError)
		set.SetOutput(stderr)
		path := set.String("manifest", "", "canonical campaign manifest")
		if err := set.Parse(arguments[1:]); err != nil {
			return err
		}
		if set.NArg() != 0 {
			return errors.New("manifest verify does not accept positional arguments")
		}
		body, err := readRequiredFile("manifest", *path)
		if err != nil {
			return err
		}
		if _, err = campaign.DecodeCanonicalManifest(body); err != nil {
			return err
		}
		return writeReport(stdout, map[string]any{"valid": true, "campaign_manifest_sha256": campaign.SHA256(body)})
	default:
		return fmt.Errorf("unknown manifest operation %q", arguments[0])
	}
}

func runAcknowledgement(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 || arguments[0] != "verify" {
		return errors.New("ack requires verify")
	}
	set := flag.NewFlagSet("ack verify", flag.ContinueOnError)
	set.SetOutput(stderr)
	manifestPath := set.String("manifest", "", "canonical campaign manifest")
	ackPath := set.String("ack", "", "canonical product acknowledgement")
	if err := set.Parse(arguments[1:]); err != nil {
		return err
	}
	if set.NArg() != 0 {
		return errors.New("ack verify does not accept positional arguments")
	}
	manifestBody, err := readRequiredFile("manifest", *manifestPath)
	if err != nil {
		return err
	}
	ackBody, err := readRequiredFile("ack", *ackPath)
	if err != nil {
		return err
	}
	manifest, err := campaign.DecodeCanonicalManifest(manifestBody)
	if err != nil {
		return err
	}
	ack, err := campaign.DecodeCanonicalAcknowledgement(ackBody)
	if err != nil {
		return err
	}
	if err = campaign.ValidateAcknowledgementAgainstManifest(ack, manifest); err != nil {
		return err
	}
	return writeReport(stdout, map[string]any{"valid": true, "product": ack.Product, "acknowledgement_sha256": campaign.SHA256(ackBody)})
}

func runAuthorization(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("authorization requires build, preimage, or verify")
	}
	switch arguments[0] {
	case "build":
		set := flag.NewFlagSet("authorization build", flag.ContinueOnError)
		set.SetOutput(stderr)
		draft := set.String("draft", "", "authorization draft JSON")
		manifest := set.String("manifest", "", "canonical campaign manifest")
		uniAck := set.String("unirouter-ack", "", "canonical UniRouter acknowledgement")
		xAck := set.String("xinsight-ack", "", "canonical X-Insight acknowledgement")
		confirmation := set.String("confirmation", "", "exact confirmation text; only its digest is emitted")
		output := set.String("out", "-", "new output path, or - for stdout")
		if err := set.Parse(arguments[1:]); err != nil {
			return err
		}
		if set.NArg() != 0 {
			return errors.New("authorization build does not accept positional arguments")
		}
		paths := map[string]string{"draft": *draft, "manifest": *manifest, "unirouter-ack": *uniAck, "xinsight-ack": *xAck, "confirmation": *confirmation}
		bodies, err := readRequiredFiles(paths)
		if err != nil {
			return err
		}
		_, canonical, err := campaign.BuildAuthorizationFromDraft(bodies["draft"], bodies["manifest"], bodies["unirouter-ack"], bodies["xinsight-ack"], bodies["confirmation"])
		if err != nil {
			return err
		}
		return writeExclusive(*output, canonical, stdout)
	case "preimage":
		set := flag.NewFlagSet("authorization preimage", flag.ContinueOnError)
		set.SetOutput(stderr)
		authorizationPath := set.String("authorization", "", "canonical authorization")
		output := set.String("out", "-", "new binary output path, or - for stdout")
		if err := set.Parse(arguments[1:]); err != nil {
			return err
		}
		if set.NArg() != 0 {
			return errors.New("authorization preimage does not accept positional arguments")
		}
		body, err := readRequiredFile("authorization", *authorizationPath)
		if err != nil {
			return err
		}
		if _, err = campaign.DecodeCanonicalAuthorization(body); err != nil {
			return err
		}
		return writeExclusive(*output, campaign.AuthorizationPreimage(body), stdout)
	case "verify":
		set := flag.NewFlagSet("authorization verify", flag.ContinueOnError)
		set.SetOutput(stderr)
		manifest := set.String("manifest", "", "canonical campaign manifest")
		uniAck := set.String("unirouter-ack", "", "canonical UniRouter acknowledgement")
		xAck := set.String("xinsight-ack", "", "canonical X-Insight acknowledgement")
		authorization := set.String("authorization", "", "canonical authorization")
		signature := set.String("signature", "", "canonical detached signature envelope")
		trustSet := set.String("trust-set", "", "canonical authorization trust set")
		if err := set.Parse(arguments[1:]); err != nil {
			return err
		}
		if set.NArg() != 0 {
			return errors.New("authorization verify does not accept positional arguments")
		}
		paths := map[string]string{"manifest": *manifest, "unirouter-ack": *uniAck, "xinsight-ack": *xAck, "authorization": *authorization, "signature": *signature, "trust-set": *trustSet}
		bodies, err := readRequiredFiles(paths)
		if err != nil {
			return err
		}
		report, err := campaign.VerifyWhole(bodies["manifest"], bodies["unirouter-ack"], bodies["xinsight-ack"], bodies["authorization"], bodies["signature"], bodies["trust-set"])
		if err != nil {
			return err
		}
		return writeReport(stdout, report)
	default:
		return fmt.Errorf("unknown authorization operation %q", arguments[0])
	}
}

func runTrustSet(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 || arguments[0] != "verify" {
		return errors.New("trust-set requires verify")
	}
	set := flag.NewFlagSet("trust-set verify", flag.ContinueOnError)
	set.SetOutput(stderr)
	path := set.String("trust-set", "", "canonical authorization trust set")
	if err := set.Parse(arguments[1:]); err != nil {
		return err
	}
	if set.NArg() != 0 {
		return errors.New("trust-set verify does not accept positional arguments")
	}
	body, err := readRequiredFile("trust-set", *path)
	if err != nil {
		return err
	}
	if _, err = campaign.DecodeCanonicalTrustSet(body); err != nil {
		return err
	}
	return writeReport(stdout, map[string]any{"valid": true, "authorization_trust_set_sha256": campaign.SHA256(body)})
}

func runRelease(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 || arguments[0] != "verify" {
		return errors.New("release requires verify")
	}
	set := flag.NewFlagSet("release verify", flag.ContinueOnError)
	set.SetOutput(stderr)
	path := set.String("release", "", "canonical release manifest")
	if err := set.Parse(arguments[1:]); err != nil {
		return err
	}
	if set.NArg() != 0 {
		return errors.New("release verify does not accept positional arguments")
	}
	body, err := readRequiredFile("release", *path)
	if err != nil {
		return err
	}
	if _, err = campaign.DecodeCanonicalRelease(body); err != nil {
		return err
	}
	return writeReport(stdout, map[string]any{"valid": true, "release_sha256": campaign.SHA256(body)})
}

func readRequiredFiles(paths map[string]string) (map[string][]byte, error) {
	bodies := make(map[string][]byte, len(paths))
	for name, path := range paths {
		body, err := readRequiredFile(name, path)
		if err != nil {
			return nil, err
		}
		bodies[name] = body
	}
	return bodies, nil
}

func readRequiredFile(name, path string) ([]byte, error) {
	if path == "" || path == "-" {
		return nil, fmt.Errorf("%s requires an explicit file path", name)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maximumInputBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	if len(body) == 0 || len(body) > maximumInputBytes {
		return nil, fmt.Errorf("%s input size is invalid", name)
	}
	return body, nil
}

func writeExclusive(path string, body []byte, stdout io.Writer) error {
	if path == "-" {
		_, err := stdout.Write(body)
		return err
	}
	directory := filepath.Dir(path)
	file, err := os.CreateTemp(directory, ".foigent-integration-control-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	if err = file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err = file.Write(body); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	// A hard link publishes the completed bytes atomically and fails if the
	// destination already exists. Rename cannot provide that no-overwrite
	// property portably on the supported Unix release targets.
	if err = os.Link(temporaryPath, path); err != nil {
		return fmt.Errorf("publish output without overwrite: %w", err)
	}
	return nil
}

func writeReport(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func usage(writer io.Writer) {
	fmt.Fprintln(writer, "foigent-integration-control: offline N1 manifest and authorization control")
	fmt.Fprintln(writer, "commands: schemas, manifest, ack, authorization, trust-set, release")
}
