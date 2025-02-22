package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Registry struct {
	v *viper.Viper
}

func NewRegistry() *Registry {
	return &Registry{
		v: viper.New(),
	}
}

func (r *Registry) LoadConfig(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		r.v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".zbittorrent")
		// configPath := filepath.Join(configDir, "config.yaml")

		r.v.AddConfigPath(configDir)
		r.v.SetConfigName("config")
		r.v.SetConfigType("yaml")
	}

	if err := r.v.ReadInConfig(); err != nil {
		fmt.Printf("Failed to read config: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	r.v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config file changed: %s\n", e.Name)
	})
	r.v.WatchConfig()

	var cfg Config

	if err := r.v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %v\n", err)
	}

	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("Error on validating config: %v", err)
	}

	return &cfg, nil
}

func (r *Registry) ConfigFile() string {
	return r.v.ConfigFileUsed()
}
