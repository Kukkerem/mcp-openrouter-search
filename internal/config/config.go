package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultModel       = "openai/gpt-5-nano"
	DefaultEngine      = "auto"
	DefaultContextSize = "medium"
	DefaultMaxResults  = 5
	DefaultTimeoutMs   = 60000
	OpenRouterEndpoint = "https://openrouter.ai/api/v1/chat/completions"
)

var (
	ValidEngines      = []string{"auto", "native", "exa", "firecrawl", "parallel"}
	ValidContextSizes = []string{"low", "medium", "high"}
)

func ResolveAPIKey(keyFile string) (string, error) {
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		return strings.TrimSpace(key), nil
	}

	path := keyFile
	if path == "" {
		path = os.Getenv("OPENROUTER_API_KEY_FILE")
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read API key from %s: %w", path, err)
		}
		if key := strings.TrimSpace(string(data)); key != "" {
			return key, nil
		}
	}

	return "", fmt.Errorf("no API key found: set OPENROUTER_API_KEY or OPENROUTER_API_KEY_FILE")
}
