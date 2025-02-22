package cmd

import (
	"fmt"
	"os"

	"github.com/al002/zbittorrent/internal/config"
	"github.com/al002/zbittorrent/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	cfgRegistry *config.Registry
	cfg         *config.Config
	log         *logger.Logger

	rootCmd = &cobra.Command{
		Use:   "zbittorrent",
		Short: "A simple bittorrent client",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("zbittorrent is starting...")
		},
	}
)

func Execute() error {
	defer func() {
		if log != nil {
			if err := log.Close(); err != nil {
				fmt.Printf("Failed to close logger: %v\n", err)
			}
		}
	}()
	return rootCmd.Execute()
}

func init() {
	cfgRegistry = config.NewRegistry()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zbittorrent/config.yaml)")
}

func initConfig() {
	var err error
	cfg, err = cfgRegistry.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err = logger.New(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Configuration loaded",
		"config_file", cfgRegistry.ConfigFile(),
		"log_dir", cfg.Log.Dir,
		"log_level", cfg.Log.Level)
}
