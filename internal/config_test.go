package internal

import (
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Ensure environment is clean for these keys
	t.Setenv("APP_LISTEN", "")
	t.Setenv("APP_TAILNET_HOST", "")
	t.Setenv("DOCKER_HOST", "")
	t.Setenv("DEFAULT_IMAGE", "")
	t.Setenv("APP_IDLE_TIMEOUT", "")
	t.Setenv("APP_SSH_PUBLIC_KEY", "")

	cfg := LoadConfig()

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr ':8080', got '%s'", cfg.ListenAddr)
	}
	if cfg.TailnetHost != "" {
		t.Errorf("expected TailnetHost '', got '%s'", cfg.TailnetHost)
	}
	if cfg.DockerHost != "unix:///var/run/docker.sock" {
		t.Errorf("expected DockerHost 'unix:///var/run/docker.sock', got '%s'", cfg.DockerHost)
	}
	if cfg.DefaultImage != "custom-opencode:latest" {
		t.Errorf("expected DefaultImage 'custom-opencode:latest', got '%s'", cfg.DefaultImage)
	}
	if cfg.IdleTimeout != 0 {
		t.Errorf("expected IdleTimeout 0, got %v", cfg.IdleTimeout)
	}
	if cfg.SSHPublicKey != "" {
		t.Errorf("expected SSHPublicKey '', got '%s'", cfg.SSHPublicKey)
	}
}

func TestLoadConfigEnvironment(t *testing.T) {
	t.Setenv("APP_LISTEN", ":9090")
	t.Setenv("APP_TAILNET_HOST", "myhost.tailnet.ts.net")
	t.Setenv("DOCKER_HOST", "tcp://1.2.3.4:2376")
	t.Setenv("DEFAULT_IMAGE", "myimage:v1")
	t.Setenv("APP_IDLE_TIMEOUT", "30m")
	t.Setenv("APP_SSH_PUBLIC_KEY", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test")
	cfg := LoadConfig()

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected ListenAddr ':9090', got '%s'", cfg.ListenAddr)
	}
	if cfg.TailnetHost != "myhost.tailnet.ts.net" {
		t.Errorf("expected TailnetHost 'myhost.tailnet.ts.net', got '%s'", cfg.TailnetHost)
	}
	if cfg.DockerHost != "tcp://1.2.3.4:2376" {
		t.Errorf("expected DockerHost 'tcp://1.2.3.4:2376', got '%s'", cfg.DockerHost)
	}
	if cfg.DefaultImage != "myimage:v1" {
		t.Errorf("expected DefaultImage 'myimage:v1', got '%s'", cfg.DefaultImage)
	}
	if cfg.IdleTimeout != 30*60*1000*1000*1000 {
		t.Errorf("expected IdleTimeout 30m, got %v", cfg.IdleTimeout)
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

func TestGetDurationEnv(t *testing.T) {
	t.Setenv("DUR_KEY", "5h")
	if d := getDurationEnv("DUR_KEY", 0); d != 5*60*60*1000*1000*1000 {
		t.Errorf("expected 5h, got %v", d)
	}
	t.Setenv("DUR_KEY", "invalid")
	if d := getDurationEnv("DUR_KEY", 1000*1000*1000); d != 1000*1000*1000 {
		t.Errorf("expected fallback 1s, got %v", d)
	}
}
