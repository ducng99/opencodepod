package internal

import (
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Ensure environment is clean for these keys
	t.Setenv("APP_LISTEN", "")
	t.Setenv("DEFAULT_IMAGE", "")
	t.Setenv("APP_SSH_PUBLIC_KEY", "")

	cfg := LoadConfig()

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr ':8080', got '%s'", cfg.ListenAddr)
	}
	if cfg.DefaultImage != "ghcr.io/ducng99/opencodepod-client:latest" {
		t.Errorf("expected DefaultImage 'ghcr.io/ducng99/opencodepod-client:latest', got '%s'", cfg.DefaultImage)
	}
	if cfg.SSHPublicKey != "" {
		t.Errorf("expected SSHPublicKey '', got '%s'", cfg.SSHPublicKey)
	}
}

func TestLoadConfigEnvironment(t *testing.T) {
	t.Setenv("APP_LISTEN", ":9090")
	t.Setenv("DEFAULT_IMAGE", "myimage:v1")
	t.Setenv("APP_SSH_PUBLIC_KEY", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test")
	cfg := LoadConfig()

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected ListenAddr ':9090', got '%s'", cfg.ListenAddr)
	}
	if cfg.DefaultImage != "myimage:v1" {
		t.Errorf("expected DefaultImage 'myimage:v1', got '%s'", cfg.DefaultImage)
	}
	if cfg.SSHPublicKey != "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test" {
		t.Errorf("expected SSHPublicKey to match, got '%s'", cfg.SSHPublicKey)
	}
}

func TestGetEnv(t *testing.T) {
	// t.Setenv overrides for the test and auto-cleans
	t.Setenv("TEST_KEY", "value")
	if v := getEnv("TEST_KEY", "default"); v != "value" {
		t.Errorf("expected 'value', got '%s'", v)
	}
	t.Setenv("TEST_KEY", "")
	if v := getEnv("TEST_KEY", "default"); v != "default" {
		t.Errorf("expected 'default', got '%s'", v)
	}
}
