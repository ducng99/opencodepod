package internal

import (
	"os"
	"time"
)

type Config struct {
	ListenAddr   string
	DockerHost   string
	DefaultImage string
	IdleTimeout  time.Duration
	SSHPublicKey string
}

func LoadConfig() *Config {
	cfg := &Config{
		ListenAddr:   getEnv("APP_LISTEN", ":8080"),
		DockerHost:   getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		DefaultImage: getEnv("DEFAULT_IMAGE", "ghcr.io/ducng99/opencodepod-client:latest"),
		IdleTimeout:  getDurationEnv("APP_IDLE_TIMEOUT", 0),
		SSHPublicKey: getEnv("APP_SSH_PUBLIC_KEY", ""),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
