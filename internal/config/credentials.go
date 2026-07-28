package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	// APIKeyEnv overrides the stored credential for CI, coding agents, and
	// temporary overrides.
	APIKeyEnv = "LINEAR_API_KEY"

	keyringService = "linear-cli"
	keyringEntry   = "api-key"
)

type CredentialSource string

const (
	CredentialSourceEnvironment CredentialSource = "environment"
	CredentialSourceKeyring     CredentialSource = "keyring"
	CredentialSourceMissing     CredentialSource = "missing"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

// ResolveAPIKey returns the Linear API key. LINEAR_API_KEY wins; otherwise the
// key is read from the OS keyring where `linear auth login` stored it.
func ResolveAPIKey() (string, CredentialSource, error) {
	if apiKey := strings.TrimSpace(os.Getenv(APIKeyEnv)); apiKey != "" {
		return apiKey, CredentialSourceEnvironment, nil
	}

	stored, err := keyring.Get(keyringService, keyringEntry)
	if err == nil {
		return stored, CredentialSourceKeyring, nil
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return "", CredentialSourceMissing, fmt.Errorf("failed to read API key from keyring: %w", err)
	}

	return "", CredentialSourceMissing, missingAPIKeyError()
}

func SaveAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("api key is required")
	}
	if err := keyring.Set(keyringService, keyringEntry, apiKey); err != nil {
		return fmt.Errorf("failed to save API key to keyring: %w", err)
	}
	return nil
}

func DeleteAPIKey() error {
	err := keyring.Delete(keyringService, keyringEntry)
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return fmt.Errorf("failed to delete API key from keyring: %w", err)
}

func missingAPIKeyError() error {
	return fmt.Errorf("%w; run 'linear auth login' or set %s", ErrAPIKeyNotFound, APIKeyEnv)
}
