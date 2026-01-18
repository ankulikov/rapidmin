package config

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var envPattern = regexp.MustCompile(`\{\{env\.([A-Za-z_][A-Za-z0-9_]*)}}`)

func Load(path string) (AppConfig, error) {
	var cfg AppConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if err := resolveProviderEnv(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func resolveProviderEnv(cfg *AppConfig) error {
	for name, provider := range cfg.Providers {
		if err := resolveProviderConfigEnv(&provider); err != nil {
			return fmt.Errorf("provider %s: %w", name, err)
		}
		cfg.Providers[name] = provider
	}

	return nil
}

func resolveProviderConfigEnv(provider *ProviderConfig) error {
	if provider.SQL != nil {
		var err error
		provider.SQL.Driver, err = resolveEnvValue(provider.SQL.Driver)
		if err != nil {
			return fmt.Errorf("driver: %w", err)
		}
		provider.SQL.DSN, err = resolveEnvValue(provider.SQL.DSN)
		if err != nil {
			return fmt.Errorf("dsn: %w", err)
		}
	}
	return nil
}

func resolveEnvValue(value string) (string, error) {
	if !envPattern.MatchString(value) {
		return value, nil
	}

	missing := map[string]struct{}{}
	resolved := envPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		envVar := parts[1]
		envValue, ok := os.LookupEnv(envVar)
		if !ok {
			missing[envVar] = struct{}{}
			return match
		}
		return envValue
	})

	if len(missing) > 0 {
		var names []string
		for name := range missing {
			names = append(names, name)
		}
		sort.Strings(names)
		return "", fmt.Errorf("missing env vars: %s", strings.Join(names, ", "))
	}

	return resolved, nil
}
