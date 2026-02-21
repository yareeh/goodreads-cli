package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goodreads-cli.yaml")
}

func SessionPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goodreads-cli-session")
}

func LoadConfig() (*Config, error) {
	// Environment variables take precedence
	email := os.Getenv("GOODREADS_EMAIL")
	password := os.Getenv("GOODREADS_PASSWORD")
	if email != "" && password != "" {
		return &Config{Email: email, Password: password}, nil
	}

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w\nCreate it with your Goodreads email and password:\n  email: you@example.com\n  password: yourpassword\nOr set GOODREADS_EMAIL and GOODREADS_PASSWORD environment variables.", ConfigPath(), err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	if cfg.Email == "" || cfg.Password == "" {
		return nil, fmt.Errorf("config file must contain 'email' and 'password' fields")
	}

	return &cfg, nil
}

// Logout removes the session file.
func Logout() error {
	path := SessionPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing session: %w", err)
	}
	return nil
}
