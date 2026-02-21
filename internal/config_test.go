package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	p := ConfigPath()
	if p == "" {
		t.Fatal("ConfigPath() returned empty string")
	}
	if filepath.Base(p) != ".goodreads-cli.yaml" {
		t.Errorf("ConfigPath() = %q, want filename .goodreads-cli.yaml", p)
	}
}

func TestSessionPath(t *testing.T) {
	p := SessionPath()
	if p == "" {
		t.Fatal("SessionPath() returned empty string")
	}
	if filepath.Base(p) != ".goodreads-cli-session" {
		t.Errorf("SessionPath() = %q, want filename .goodreads-cli-session", p)
	}
}

func TestLoadConfigValid(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configContent := `email: test@example.com
password: secret123`
	os.WriteFile(filepath.Join(tmpDir, ".goodreads-cli.yaml"), []byte(configContent), 0600)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "test@example.com")
	}
	if cfg.Password != "secret123" {
		t.Errorf("Password = %q, want %q", cfg.Password, "secret123")
	}
}

func TestLoadConfigMissing(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfigMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name    string
		content string
	}{
		{"missing email", "password: secret"},
		{"missing password", "email: test@example.com"},
		{"empty values", "email: \"\"\npassword: \"\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(filepath.Join(tmpDir, ".goodreads-cli.yaml"), []byte(tt.content), 0600)
			_, err := LoadConfig()
			if err == nil {
				t.Errorf("expected error for config with %s", tt.name)
			}
		})
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	os.WriteFile(filepath.Join(tmpDir, ".goodreads-cli.yaml"), []byte(":::invalid:::"), 0600)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
