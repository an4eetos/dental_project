package config

import infraconfig "github.com/anuarkuanysh/dental_project/infra/config"

// Config re-exports infra configuration for legacy imports.
type Config = infraconfig.Config

// Load reads configuration from environment variables.
func Load() (Config, error) {
	return infraconfig.Load()
}
