package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte(" test-token \n"), 0o600); err != nil {
		t.Fatal(err)
	}

	token, err := ReadToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if token != "test-token" {
		t.Fatalf("expected trimmed token, got %q", token)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte("abc123"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OBOT_STORAGE_URL", "https://obot-storage.default.svc:8443")
	t.Setenv("OBOT_STORAGE_TOKEN_FILE", path)
	t.Setenv("MCP_RUNTIME_NAMESPACE", "obot-mcp")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ObotStorageURL != "https://obot-storage.default.svc:8443" {
		t.Fatalf("unexpected storage url %q", cfg.ObotStorageURL)
	}
	if cfg.ObotStorageTokenFile != path {
		t.Fatalf("unexpected token file %q", cfg.ObotStorageTokenFile)
	}
	if cfg.MCPRuntimeNamespace != "obot-mcp" {
		t.Fatalf("unexpected runtime namespace %q", cfg.MCPRuntimeNamespace)
	}
}
