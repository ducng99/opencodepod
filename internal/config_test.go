package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfigFrom(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr ':8080', got '%s'", cfg.ListenAddr)
	}
	if cfg.DefaultImage != "ghcr.io/ducng99/opencodepod-client:latest" {
		t.Errorf("expected DefaultImage 'ghcr.io/ducng99/opencodepod-client:latest', got '%s'", cfg.DefaultImage)
	}
	if cfg.SSHPublicKey != "" {
		t.Errorf("expected SSHPublicKey '', got '%s'", cfg.SSHPublicKey)
	}
	if cfg.Git.Auth.SSHKey != "" {
		t.Errorf("expected Git.Auth.SSHKey '', got '%s'", cfg.Git.Auth.SSHKey)
	}
	if cfg.Git.Auth.SSHKeyPath != "/home/coder/.ssh/id_ed25519" {
		t.Errorf("expected Git.Auth.SSHKeyPath '/home/coder/.ssh/id_ed25519', got '%s'", cfg.Git.Auth.SSHKeyPath)
	}
}

func TestLoadConfigFromJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{
		"listen_addr": ":9090",
		"default_image": "myimage:v1",
		"ssh_public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test",
		"git": {
			"auth": {
				"ssh_key": "mykey",
				"ssh_key_path": "/custom/key"
			}
		}
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected ListenAddr ':9090', got '%s'", cfg.ListenAddr)
	}
	if cfg.DefaultImage != "myimage:v1" {
		t.Errorf("expected DefaultImage 'myimage:v1', got '%s'", cfg.DefaultImage)
	}
	if cfg.SSHPublicKey != "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test" {
		t.Errorf("expected SSHPublicKey to match, got '%s'", cfg.SSHPublicKey)
	}
	if cfg.Git.Auth.SSHKey != "mykey" {
		t.Errorf("expected Git.Auth.SSHKey 'mykey', got '%s'", cfg.Git.Auth.SSHKey)
	}
	if cfg.Git.Auth.SSHKeyPath != "/custom/key" {
		t.Errorf("expected Git.Auth.SSHKeyPath '/custom/key', got '%s'", cfg.Git.Auth.SSHKeyPath)
	}
}

func TestLoadConfigJSONPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{
		"listen_addr": ":7070"
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":7070" {
		t.Errorf("expected ListenAddr ':7070', got '%s'", cfg.ListenAddr)
	}
	if cfg.DefaultImage != "ghcr.io/ducng99/opencodepod-client:latest" {
		t.Errorf("expected DefaultImage default, got '%s'", cfg.DefaultImage)
	}
}
