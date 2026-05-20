package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

type Mount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
}

type GitAuthConfig struct {
	SSHKey     string `json:"ssh_key"`
	SSHKeyPath string `json:"ssh_key_path"`
}

type GitConfig struct {
	Auth      GitAuthConfig `json:"auth"`
	UserName  string        `json:"user_name"`
	UserEmail string        `json:"user_email"`
}

type Config struct {
	ListenAddr   string    `json:"listen_addr"`
	DefaultImage string    `json:"default_image"`
	SSHPublicKey string    `json:"ssh_public_key"`
	Mounts       []Mount   `json:"mounts"`
	Git          GitConfig `json:"git"`
}

const defaultConfigPath = "config.json"

var filePlaceholderRe = regexp.MustCompile(`^\{file:(.+)\}$`)

func LoadConfig() (*Config, error) {
	return loadConfigFrom(defaultConfigPath)
}

func loadConfigFrom(path string) (*Config, error) {
	cfg := &Config{
		ListenAddr:   ":8080",
		DefaultImage: "ghcr.io/ducng99/opencodepod-client:latest",
		SSHPublicKey: "",
		Mounts:       []Mount{},
		Git: GitConfig{
			Auth: GitAuthConfig{
				SSHKeyPath: "/home/coder/.ssh/id_ed25519",
			},
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	configDir := filepath.Dir(path)
	if err := expandPlaceholders(cfg, configDir); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func expandPlaceholders(v interface{}, configDir string) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return expandValue(rv, configDir)
}

func expandValue(v reflect.Value, configDir string) error {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if m := filePlaceholderRe.FindStringSubmatch(s); m != nil {
			filePath := m[1]
			if !filepath.IsAbs(filePath) {
				filePath = filepath.Join(configDir, filePath)
			}
			content, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			v.SetString(string(content))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := expandValue(v.Field(i), configDir); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := expandValue(v.Index(i), configDir); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			if err := expandValue(v.Elem(), configDir); err != nil {
				return err
			}
		}
	}
	return nil
}
