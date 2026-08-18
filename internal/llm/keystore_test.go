package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

// testCredentialsDir overrides credentialsFilePath for isolated tests.
// Each test that touches the filesystem should use t.TempDir().
func setupTestCredentials(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()

	// Override the credentials file path for tests by writing to a
	// known location under t.TempDir(). Since credentialsFilePath()
	// uses os.UserHomeDir(), we manipulate HOME for the test.
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)

	// Create the config directory structure
	configDir := filepath.Join(dir, ".config", "ais")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create test config dir: %v", err)
	}

	cleanup = func() {
		os.Setenv("HOME", origHome)
	}
	return dir, cleanup
}

func writeTestCredentials(t *testing.T, dir, key string) {
	t.Helper()
	path := filepath.Join(dir, ".config", "ais", "credentials.json")
	creds := Credentials{AnthropicAPIKey: key}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		t.Fatalf("marshal test credentials: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test credentials: %v", err)
	}
}

func TestGetAPIKey_EnvVar(t *testing.T) {
	testKey := "sk-ant-test-env-key-123"
	t.Setenv("ANTHROPIC_API_KEY", testKey)

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != testKey {
		t.Errorf("expected key from env var, got different value")
	}
}

func TestGetAPIKey_EnvVarPriority(t *testing.T) {
	// Env var should take priority over credentials file
	envKey := "sk-ant-env-priority-key"
	fileKey := "sk-ant-file-key-456"

	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	writeTestCredentials(t, dir, fileKey)
	t.Setenv("ANTHROPIC_API_KEY", envKey)

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != envKey {
		t.Errorf("expected env var key to take priority over file key")
	}
}

func TestGetAPIKey_CredentialsFile(t *testing.T) {
	// Clear env var so file fallback is used
	t.Setenv("ANTHROPIC_API_KEY", "")

	// Clear keychain to ensure file fallback is tested
	keyring.Delete(keychainService, keychainUser)

	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	testKey := "sk-ant-file-fallback-key"
	writeTestCredentials(t, dir, testKey)

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != testKey {
		t.Errorf("expected key from credentials file")
	}
}

func TestGetAPIKey_NoKeyConfigured(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Clear keychain to ensure no key is found
	keyring.Delete(keychainService, keychainUser)

	_, err := GetAPIKey()
	if err == nil {
		t.Error("expected error when no key is configured")
	}
}

func TestSetAPIKey_InvalidFormat(t *testing.T) {
	err := SetAPIKey("invalid-key-format")
	if err == nil {
		t.Error("expected error for invalid key format")
	}
}

func TestSetAPIKey_ValidFormat(t *testing.T) {
	_, cleanup := setupTestCredentials(t)
	defer cleanup()

	// Clear env var
	t.Setenv("ANTHROPIC_API_KEY", "")

	testKey := "sk-ant-valid-test-key-789"
	err := SetAPIKey(testKey)
	if err != nil {
		t.Fatalf("SetAPIKey failed: %v", err)
	}

	// Clean up keychain entry after test to avoid polluting other tests
	t.Cleanup(func() {
		keyring.Delete(keychainService, keychainUser)
	})

	// Verify the key was stored (either keychain or file)
	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey after SetAPIKey failed: %v", err)
	}
	if key != testKey {
		t.Errorf("stored key does not match")
	}
}

func TestSetAPIKey_CredentialsFilePermissions(t *testing.T) {
	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	t.Setenv("ANTHROPIC_API_KEY", "")

	testKey := "sk-ant-permissions-test-key"

	// Write directly via the file fallback to ensure we test file creation
	err := writeCredentialsFile(testKey)
	if err != nil {
		t.Fatalf("writeCredentialsFile failed: %v", err)
	}

	path := filepath.Join(dir, ".config", "ais", "credentials.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials file: %v", err)
	}

	// Verify 0600 permissions (owner read/write only)
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected credentials file permissions 0600, got %o", perm)
	}
}

func TestHasAPIKey_WithEnvVar(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-has-key-test")

	if !HasAPIKey() {
		t.Error("HasAPIKey should return true when env var is set")
	}
}

func TestHasAPIKey_WithCredentialsFile(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	writeTestCredentials(t, dir, "sk-ant-has-key-file-test")

	if !HasAPIKey() {
		t.Error("HasAPIKey should return true when credentials file exists")
	}
}

func TestHasAPIKey_NoKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Ensure no key in keychain from previous tests
	keyring.Delete(keychainService, keychainUser)

	if HasAPIKey() {
		t.Error("HasAPIKey should return false when no key is configured")
	}
}

func TestDeleteAPIKey_RemovesCredentialsFile(t *testing.T) {
	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	testKey := "sk-ant-delete-test-key"
	writeTestCredentials(t, dir, testKey)

	path := filepath.Join(dir, ".config", "ais", "credentials.json")

	// Verify file exists before delete
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("credentials file should exist before delete")
	}

	err := DeleteAPIKey()
	if err != nil {
		t.Fatalf("DeleteAPIKey failed: %v", err)
	}

	// Verify file is removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("credentials file should be removed after DeleteAPIKey")
	}
}

func TestDeleteAPIKey_NoKeyIsNotError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// DeleteAPIKey should not error when no key exists
	err := DeleteAPIKey()
	if err != nil {
		t.Errorf("DeleteAPIKey should not error when no key exists: %v", err)
	}
}

func TestCredentialsFileFormat(t *testing.T) {
	dir, cleanup := setupTestCredentials(t)
	defer cleanup()

	testKey := "sk-ant-format-test-key"
	if err := writeCredentialsFile(testKey); err != nil {
		t.Fatalf("writeCredentialsFile failed: %v", err)
	}

	path := filepath.Join(dir, ".config", "ais", "credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read credentials file: %v", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatalf("credentials file is not valid JSON: %v", err)
	}
	if creds.AnthropicAPIKey != testKey {
		t.Errorf("credentials file key mismatch")
	}
}

func TestKeyFormatValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{name: "valid key", key: "sk-ant-api03-abc123", wantErr: false},
		{name: "empty key", key: "", wantErr: true},
		{name: "no prefix", key: "abc123", wantErr: true},
		{name: "wrong prefix", key: "sk-test-123", wantErr: true},
		{name: "partial prefix", key: "sk-ant", wantErr: true},
		{name: "valid prefix with dash", key: "sk-ant-x", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We only test format validation, so we set HOME to a temp dir
			// to avoid polluting real keychain/files.
			dir := t.TempDir()
			t.Setenv("HOME", dir)
			os.MkdirAll(filepath.Join(dir, ".config", "ais"), 0o755)

			err := SetAPIKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAPIKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}
