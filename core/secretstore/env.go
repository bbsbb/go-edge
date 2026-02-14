package secretstore

import (
	"fmt"
	"os"
	"strings"
)

// EnvService resolves secrets from environment variables.
// It converts the secret name to uppercase and replaces hyphens with
// underscores (e.g., "db-password" → "DB_PASSWORD"). An optional prefix
// scopes lookups (e.g., prefix "SECRET" → "SECRET_DB_PASSWORD").
type EnvService struct {
	prefix string
}

// NewEnvService creates a secret store that resolves from environment variables.
// If prefix is non-empty, it's prepended to the variable name with an underscore.
func NewEnvService(prefix string) *EnvService {
	return &EnvService{prefix: prefix}
}

func (s *EnvService) GetSecretValue(secretName string) (string, error) {
	envKey := strings.ToUpper(strings.ReplaceAll(secretName, "-", "_"))
	if s.prefix != "" {
		envKey = s.prefix + "_" + envKey
	}

	value, ok := os.LookupEnv(envKey)
	if !ok {
		return "", fmt.Errorf("environment variable %s not set for secret %q", envKey, secretName)
	}

	return value, nil
}
