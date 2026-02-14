package versions

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/bbsbb/go-edge/sweetshop/internal/config"
)

const AppSchemaName = "app_sweetshop"

func getAppConfiguration(ctx context.Context) (*config.AppConfiguration, error) {
	_, filename, _, _ := runtime.Caller(0)
	configDirectory := path.Join(path.Dir(filename), "../../..", "resources", "config")

	cfg, err := config.NewAppConfiguration(ctx, configDirectory)
	if err != nil {
		return nil, fmt.Errorf("load app configuration: %w", err)
	}

	return cfg, nil
}

func getMigrateConfiguration(ctx context.Context) (*config.MigrateConfiguration, error) {
	_, filename, _, _ := runtime.Caller(0)
	configDirectory := path.Join(path.Dir(filename), "../../..", "resources", "config", "migrate")

	cfg, err := config.NewMigrateConfiguration(ctx, configDirectory)
	if err != nil {
		return nil, fmt.Errorf("load migrate configuration: %w", err)
	}

	return cfg, nil
}
