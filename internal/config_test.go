package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestLoadConfigPlaceholderHappyPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	keyPath := filepath.Join(dir, "ssh_key.txt")

	if err := os.WriteFile(keyPath, []byte("key-from-file"), 0644); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	content := `{
		"ssh_public_key": "{file:ssh_key.txt}"
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SSHPublicKey != "key-from-file" {
		t.Errorf("expected SSHPublicKey 'key-from-file', got '%s'", cfg.SSHPublicKey)
	}
}

func TestLoadConfigPlaceholderMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{
		"ssh_public_key": "{file:missing_key.txt}"
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFrom(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfigPlaceholderNestedStruct(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	keyPath := filepath.Join(dir, "git_key.txt")

	if err := os.WriteFile(keyPath, []byte("git-key-content"), 0644); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	content := `{
		"git": {
			"auth": {
				"ssh_key": "{file:git_key.txt}"
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

	if cfg.Git.Auth.SSHKey != "git-key-content" {
		t.Errorf("expected Git.Auth.SSHKey 'git-key-content', got '%s'", cfg.Git.Auth.SSHKey)
	}
}

func TestLoadConfigPlaceholderSlice(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	mountSourcePath := filepath.Join(dir, "mount_source.txt")

	if err := os.WriteFile(mountSourcePath, []byte("/host/path"), 0644); err != nil {
		t.Fatalf("write mount source file: %v", err)
	}

	content := `{
		"mounts": [
			{"source": "{file:mount_source.txt}", "target": "/container/path"}
		]
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(cfg.Mounts))
	}
	if cfg.Mounts[0].Source != "/host/path" {
		t.Errorf("expected mount source '/host/path', got '%s'", cfg.Mounts[0].Source)
	}
	if cfg.Mounts[0].Target != "/container/path" {
		t.Errorf("expected mount target '/container/path', got '%s'", cfg.Mounts[0].Target)
	}
}

func TestLoadConfigPlaceholderAbsolutePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	keyPath := filepath.Join(dir, "ssh_key.txt")

	if err := os.WriteFile(keyPath, []byte("abs-path-key"), 0644); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	content := `{
		"ssh_public_key": "{file:` + filepath.ToSlash(keyPath) + `}"
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SSHPublicKey != "abs-path-key" {
		t.Errorf("expected SSHPublicKey 'abs-path-key', got '%s'", cfg.SSHPublicKey)
	}
}

func TestLoadConfigHosts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{
		"hosts": {
			"myapp.local": "192.168.1.100",
			"db.local": "10.0.0.5"
		}
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(cfg.Hosts))
	}
	if cfg.Hosts["myapp.local"] != "192.168.1.100" {
		t.Errorf("expected myapp.local -> 192.168.1.100, got %s", cfg.Hosts["myapp.local"])
	}
	if cfg.Hosts["db.local"] != "10.0.0.5" {
		t.Errorf("expected db.local -> 10.0.0.5, got %s", cfg.Hosts["db.local"])
	}
}

func TestLoadConfigHostsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hosts should be nil when not specified in JSON
	if cfg.Hosts != nil {
		t.Errorf("expected Hosts to be nil when not specified, got %v", cfg.Hosts)
	}
}
