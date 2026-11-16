package cli

import (
	"github.com/Thenlie/clask/internal/config"
)

// configPath returns the explicit --config path or the XDG default.
func configPath() (string, error) {
	if flags.configPath != "" {
		return config.ExpandHome(flags.configPath)
	}
	return config.DefaultPath()
}

// loadConfig reads the registry, returning defaults when none exists yet.
func loadConfig() (*config.Config, string, error) {
	path, err := configPath()
	if err != nil {
		return nil, "", err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}
