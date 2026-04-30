package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAPIKey_EnvVar(t *testing.T) {
	os.Setenv("OPENROUTER_API_KEY", "test-key-from-env")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	key, err := ResolveAPIKey("")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "test-key-from-env" {
		t.Errorf("key = %q", key)
	}
}

func TestResolveAPIKey_File(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")

	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	os.WriteFile(keyFile, []byte("  test-key-from-file  \n"), 0644)

	os.Setenv("OPENROUTER_API_KEY_FILE", keyFile)
	defer os.Unsetenv("OPENROUTER_API_KEY_FILE")

	key, err := ResolveAPIKey("")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "test-key-from-file" {
		t.Errorf("key = %q", key)
	}
}

func TestResolveAPIKey_FileArg(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY_FILE")

	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	os.WriteFile(keyFile, []byte("key-from-arg"), 0644)

	key, err := ResolveAPIKey(keyFile)
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "key-from-arg" {
		t.Errorf("key = %q", key)
	}
}

func TestResolveAPIKey_Missing(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY_FILE")

	_, err := ResolveAPIKey("")
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestResolveAPIKey_EnvTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	os.WriteFile(keyFile, []byte("file-key"), 0644)

	os.Setenv("OPENROUTER_API_KEY", "env-key")
	os.Setenv("OPENROUTER_API_KEY_FILE", keyFile)
	defer os.Unsetenv("OPENROUTER_API_KEY")
	defer os.Unsetenv("OPENROUTER_API_KEY_FILE")

	key, err := ResolveAPIKey("")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "env-key" {
		t.Errorf("key = %q, want env-key to take precedence", key)
	}
}
