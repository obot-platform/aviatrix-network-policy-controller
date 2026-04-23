package config

import (
	"fmt"
	"os"
	"strings"
)

const defaultTokenFile = "/var/run/secrets/obot-network-policy-provider/apiKey"

type Config struct {
	ObotStorageURL       string
	ObotStorageTokenFile string
	MCPRuntimeNamespace  string
}

func Load() (Config, error) {
	cfg := Config{
		ObotStorageURL:       strings.TrimSpace(os.Getenv("OBOT_STORAGE_URL")),
		ObotStorageTokenFile: strings.TrimSpace(os.Getenv("OBOT_STORAGE_TOKEN_FILE")),
		MCPRuntimeNamespace:  strings.TrimSpace(os.Getenv("MCP_RUNTIME_NAMESPACE")),
	}

	if cfg.ObotStorageTokenFile == "" {
		cfg.ObotStorageTokenFile = defaultTokenFile
	}
	if cfg.ObotStorageURL == "" {
		return Config{}, fmt.Errorf("OBOT_STORAGE_URL is required")
	}
	if cfg.MCPRuntimeNamespace == "" {
		return Config{}, fmt.Errorf("MCP_RUNTIME_NAMESPACE is required")
	}
	if _, err := ReadToken(cfg.ObotStorageTokenFile); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ReadToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read Obot storage token file %q: %w", path, err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("Obot storage token file %q is empty", path)
	}
	return token, nil
}
