package config

import (
	"fmt"
	"strings"
)

func ValidateSecurity(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	enforce := cfg.Server.Environment == "production" || cfg.Security.RequireStrongSecrets
	if !enforce {
		return nil
	}

	if strings.TrimSpace(cfg.JWT.PrivateKey) == "" || strings.TrimSpace(cfg.JWT.PublicKey) == "" || strings.TrimSpace(cfg.JWT.RefreshSecret) == "" {
		return fmt.Errorf("jwt signing material and refresh secret are required")
	}
	if strings.TrimSpace(cfg.Security.InternalToken) == "" {
		return fmt.Errorf("internal auth token is required")
	}
	if strings.TrimSpace(cfg.Database.Password) == "" || cfg.Database.Password == "postgres" {
		return fmt.Errorf("database password must be non-default")
	}
	if strings.TrimSpace(cfg.Redis.Password) == "" && cfg.Server.Environment == "production" {
		return fmt.Errorf("redis password is required in production")
	}
	if cfg.Security.MTLSRequired {
		if strings.TrimSpace(cfg.Server.TLSCertFile) == "" || strings.TrimSpace(cfg.Server.TLSKeyFile) == "" || strings.TrimSpace(cfg.Server.ClientCAFile) == "" {
			return fmt.Errorf("mTLS requires TLS cert, key, and client CA files")
		}
	}
	return nil
}
