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

	// Clear env vars so they don't mask the missing file
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Unsetenv("GOODREADS_EMAIL")
	os.Unsetenv("GOODREADS_PASSWORD")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		}
	}()

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

	// Clear env vars
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Unsetenv("GOODREADS_EMAIL")
	os.Unsetenv("GOODREADS_PASSWORD")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		}
	}()

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

func TestLoadConfigFromEnvVars(t *testing.T) {
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Setenv("GOODREADS_EMAIL", "env@example.com")
	os.Setenv("GOODREADS_PASSWORD", "envpass")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		} else {
			os.Unsetenv("GOODREADS_EMAIL")
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		} else {
			os.Unsetenv("GOODREADS_PASSWORD")
		}
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Email != "env@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "env@example.com")
	}
	if cfg.Password != "envpass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "envpass")
	}
}

func TestLoadConfigEnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Write a config file
	os.WriteFile(filepath.Join(tmpDir, ".goodreads-cli.yaml"), []byte("email: file@example.com\npassword: filepass"), 0600)

	// Set env vars — should take precedence
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Setenv("GOODREADS_EMAIL", "env@example.com")
	os.Setenv("GOODREADS_PASSWORD", "envpass")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		} else {
			os.Unsetenv("GOODREADS_EMAIL")
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		} else {
			os.Unsetenv("GOODREADS_PASSWORD")
		}
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Email != "env@example.com" {
		t.Errorf("Email = %q, want env value", cfg.Email)
	}
}

func TestLogout(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create fake session
	os.WriteFile(SessionPath(), []byte("cookies"), 0600)

	if err := Logout(); err != nil {
		t.Fatalf("Logout() error: %v", err)
	}
	if _, err := os.Stat(SessionPath()); !os.IsNotExist(err) {
		t.Error("session file still exists after Logout()")
	}
}

func TestLogoutNoSession(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Should not error when no session file exists
	if err := Logout(); err != nil {
		t.Fatalf("Logout() error: %v", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Clear env vars
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Unsetenv("GOODREADS_EMAIL")
	os.Unsetenv("GOODREADS_PASSWORD")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		}
	}()

	os.WriteFile(filepath.Join(tmpDir, ".goodreads-cli.yaml"), []byte(":::invalid:::"), 0600)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
