package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppUsername string
	AppPassword string
	JWTSecret   string
	AppPort     int
	StorageType string

	CephEndpoint  string
	CephAccessKey string
	CephSecretKey string
	CephRegion    string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppUsername:   os.Getenv("APP_USERNAME"),
		AppPassword:   os.Getenv("APP_PASSWORD"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		StorageType:   getEnvDefault("STORAGE_TYPE", "ceph"),
		CephEndpoint:  os.Getenv("CEPH_ENDPOINT"),
		CephAccessKey: os.Getenv("CEPH_ACCESS_KEY"),
		CephSecretKey: os.Getenv("CEPH_SECRET_KEY"),
		CephRegion:    getEnvDefault("CEPH_REGION", "default"),
	}

	port, err := strconv.Atoi(getEnvDefault("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}
	cfg.AppPort = port

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"APP_USERNAME":    c.AppUsername,
		"APP_PASSWORD":    c.AppPassword,
		"JWT_SECRET":      c.JWTSecret,
		"CEPH_ENDPOINT":   c.CephEndpoint,
		"CEPH_ACCESS_KEY": c.CephAccessKey,
		"CEPH_SECRET_KEY": c.CephSecretKey,
	}
	for name, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
