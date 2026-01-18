package rapidmin

import (
	"context"
	"fmt"

	"github.com/ankulikov/rapidmin/config"
	"github.com/ankulikov/rapidmin/providers"
	sqlprovider "github.com/ankulikov/rapidmin/providers/sql"
	"github.com/ankulikov/rapidmin/server"
)

func NewServer(cfg config.AppConfig) (*server.Server, error) {
	registry, err := buildProviders(cfg)
	if err != nil {
		return nil, err
	}

	return server.New(cfg, registry)
}

func buildProviders(cfg config.AppConfig) (providers.Registry, error) {
	registry := providers.Registry{}

	for name, providerConfig := range cfg.Providers {
		if providerConfig.SQL == nil {
			return nil, fmt.Errorf("provider %s missing sql config", name)
		}
		sqlProvider := sqlprovider.New()
		if err := sqlProvider.Init(context.Background(), name, providerConfig); err != nil {
			return nil, fmt.Errorf("failed to initialize sql provider %s: %w", name, err)
		}

		registry[name] = sqlProvider
	}

	return registry, nil
}
