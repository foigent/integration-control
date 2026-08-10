package campaign

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCapabilityBoundaryHasNoProductOrSigningAdapters(t *testing.T) {
	root := filepath.Join("..", "..")
	forbiddenImports := map[string]bool{
		"database/sql": true,
		"net":          true,
		"net/http":     true,
		"net/rpc":      true,
		"os/exec":      true,
		"plugin":       true,
		"syscall":      true,
	}
	for _, directory := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imported := range parsed.Imports {
				name, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return err
				}
				if forbiddenImports[name] {
					t.Errorf("%s imports forbidden capability %q", path, name)
				}
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			source := string(body)
			for _, forbidden := range []string{
				"ed25519." + "Sign(",
				"ed25519." + "GenerateKey(",
				"BEGIN " + "PRIVATE KEY",
				"postgres" + "://",
				"redis" + "://",
			} {
				if strings.Contains(source, forbidden) {
					t.Errorf("%s contains forbidden capability marker %q", path, forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRepositoryContainsNoPrivateKeyMaterial(t *testing.T) {
	root := filepath.Join("..", "..")
	pemPrivateMarker := "-----BEGIN " + "PRIVATE KEY-----"
	pemOpenSSHMarker := "-----BEGIN OPENSSH " + "PRIVATE KEY-----"
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, pemPrivateMarker) ||
				strings.Contains(line, pemOpenSSHMarker) {
				t.Errorf("%s contains private-key material", path)
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestModuleUsesOnlyPinnedStandardLibraryToolchain(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "module github.com/foigent/integration-control\n\ngo 1.25.0\n\ntoolchain go1.25.12\n"
	if string(body) != expected {
		t.Fatalf("go.mod drifted from the accepted toolchain-only module:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join("..", "..", "go.sum")); !os.IsNotExist(err) {
		t.Fatal("standard-library-only module must not have go.sum")
	}
}
