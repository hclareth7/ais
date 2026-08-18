package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keychainService = "ais"
	keychainUser    = "api-key"
	keyPrefix       = "sk-ant-"
)

// GetAPIKey resolves the Anthropic API key using the following priority order:
//  1. ANTHROPIC_API_KEY environment variable
//  2. OS keychain via go-keyring (service: "ais", user: "api-key")
//  3. Fallback credentials file: ~/.config/ais/credentials.json (0600 permissions)
//
// Returns the first non-empty key found, or an error if no key is available.
// The key value is never logged or included in error messages.
func GetAPIKey() (string, error) {
	// 1. Environment variable
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// 2. OS keychain
	key, err := keyring.Get(keychainService, keychainUser)
	if err == nil && key != "" {
		return key, nil
	}

	// 3. Credentials file fallback
	key, err = readCredentialsFile()
	if err == nil && key != "" {
		return key, nil
	}

	return "", fmt.Errorf("no API key configured: set ANTHROPIC_API_KEY, use OS keychain, or configure via settings")
}

// SetAPIKey stores the API key securely. It attempts the OS keychain first,
// falling back to the credentials file (~/.config/ais/credentials.json)
// with 0600 permissions when the keychain is unavailable.
//
// The key format is validated before storage: it must start with "sk-ant-".
// The key is never logged or included in error messages.
func SetAPIKey(key string) error {
	if !strings.HasPrefix(key, keyPrefix) {
		return fmt.Errorf("invalid API key format")
	}

	// Try OS keychain first
	err := keyring.Set(keychainService, keychainUser, key)
	if err == nil {
		return nil
	}

	// Fallback: write to credentials file
	return writeCredentialsFile(key)
}

// HasAPIKey returns true if an API key is available from any source in the
// resolution chain. It never exposes the key value.
func HasAPIKey() bool {
	key, err := GetAPIKey()
	return err == nil && key != ""
}

// DeleteAPIKey removes the API key from all storage locations:
// OS keychain and the credentials file. Errors from individual
// deletions are collected but do not prevent other deletions.
func DeleteAPIKey() error {
	var errs []string

	// Delete from keychain (ignore ErrNotFound)
	if err := keyring.Delete(keychainService, keychainUser); err != nil && err != keyring.ErrNotFound {
		errs = append(errs, fmt.Sprintf("keychain: %v", err))
	}

	// Delete credentials file
	if err := deleteCredentialsFile(); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Sprintf("credentials file: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("delete API key: %s", strings.Join(errs, "; "))
	}
	return nil
}

// credentialsFilePath returns the path to the credentials file.
// This is separate from config.json to maintain the security boundary:
// config.json is 0644 (world-readable), credentials.json is 0600 (owner-only).
// Returns an error if the home directory cannot be determined — falling back
// to os.TempDir() would expose credentials on shared systems.
func credentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "ais", "credentials.json"), nil
}

// readCredentialsFile reads the API key from the credentials file.
func readCredentialsFile() (string, error) {
	path, err := credentialsFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read credentials file: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parse credentials file: %w", err)
	}

	return creds.AnthropicAPIKey, nil
}

// writeCredentialsFile writes the API key to the credentials file with
// 0600 permissions. Creates the parent directory if it does not exist.
func writeCredentialsFile(key string) error {
	path, err := credentialsFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}

	creds := Credentials{AnthropicAPIKey: key}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials file: %w", err)
	}

	return nil
}

// deleteCredentialsFile removes the credentials file from disk.
func deleteCredentialsFile() error {
	path, err := credentialsFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
